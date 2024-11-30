from plexapi.myplex import PlexServer
from plexapi.client import PlexClient, TIMEOUT
from flask import Flask
import random
from flask import request
import config
from flask import Flask, jsonify
from flask_cors import CORS
import os
import logging

# Initialize Flask app
app = Flask(__name__)
allowed_origin = os.getenv('CORS_ORIGIN', 'http://localhost:6969')
CORS(app, resources={r"/*": {"origins": allowed_origin}})

# Configure logging
log_dir = os.path.join(os.path.dirname(os.path.abspath(__file__)), 'logs')
os.makedirs(log_dir, exist_ok=True)

logging.basicConfig(
    level=logging.INFO,
    format='%(asctime)s - %(name)s - %(levelname)s - %(message)s',
    handlers=[
        logging.FileHandler(os.path.join(log_dir, 'app.log')),
        logging.StreamHandler()
    ]
)

logger = logging.getLogger(__name__)

def login():
    baseurl = 'http://' + config.PLEX_IP + ':' + config.PLEX_PORT
    token = 'Y7fU6x3PPqr8A-P3WEjq'
    plex = PlexServer(baseurl, token)
    return plex

def devices():
    plex = login()

    avaliable_devices = ""
    devices = plex.systemDevices()
    for device in devices:
        avaliable_devices += str(device) + "\n"
    print(avaliable_devices)

@app.route("/movies", methods=['GET'])
def get_movies():
    movies_list = []

    plex = login()
    movies = plex.library.section('Movies')
    for video in movies.search(unwatched=True):
        movies_list.append({
            "title": video.title,
            "tagline": video.tagline,
            "summary": video.summary,
            "year": video.year,
            "duration": video.duration,
            "rating": video.rating,
            "genres": video.genres,
            "actors": [actor.tag for actor in video.actors]
        })

    response = jsonify({'movies': len(movies_list)})
    response.headers.add('Access-Control-Allow-Origin', '*')
    return response

@app.route("/shows", methods=['GET'])
def get_shows():
    shows_list = []

    plex = login()
    shows = plex.library.section('TV Shows')
    for show in shows.search(unwatched=True):
        show_details = {
            "title": show.title,
            "summary": show.summary,
            "year": show.year,
            "rating": show.rating,
            "genres": show.genres,
            "episodes": []
        }
        for episode in show.episodes():
            episode_details = {
                "title": episode.title,
                "season": episode.seasonNumber,
                "episode": episode.index,
                "summary": episode.summary,
                "duration": episode.duration,
                "rating": episode.rating
            }
            show_details["episodes"].append(episode_details)
        shows_list.append(show_details)

    response = jsonify({'shows': len(shows_list)})
    response.headers.add('Access-Control-Allow-Origin', '*')
    return response


@app.route("/get-random-media", methods=['POST'])
def get_random_media():
    try:
        data = request.json
        count = int(data.get('number', 1))
        media_type = data.get('type', 'movie')

        logger.info(f"Random {media_type} request - count: {count}")

        plex = login()
        library = plex.library.section('Movies' if media_type == 'movie' else 'TV Shows')
        available_media = library.search(unwatched=True)

        if not available_media:
            return jsonify({'error': 'No unwatched media available'}), 404

        selected_items = random.sample(available_media, min(count, len(available_media)))
        queue = plex.createPlayQueue(selected_items[0])

        queued_items = [selected_items[0].title]
        for item in selected_items[1:]:
            queue.addItem(item)
            queued_items.append(item.title)

        client = plex.client("Ryzen-Win")
        client.playMedia(queue)

        response = {
            'queue_length': len(queued_items),
            'items': queued_items
        }
        logger.info(f"Queued items: {response}")
        return jsonify(response)

    except Exception as e:
        logger.error(f"Error in get_random_media: {str(e)}", exc_info=True)
        return jsonify({'error': str(e)}), 500


@app.route("/queue-specific", methods=['POST'])
def queue_specific():
    try:
        data = request.json
        logger.info(f"Received queue request: {data}")

        media_type = data.get('type', 'movie')
        titles = data.get('items', [])

        if not titles:
            return jsonify({'error': 'No titles provided'}), 400

        plex = login()
        library = plex.library.section('Movies' if media_type == 'movie' else 'TV Shows')

        queued_items = []
        first_item = None

        for search_term in titles:
            logger.info(f"Searching for: {search_term}")
            matches = library.search(title=search_term)
            logger.info(f"Found {len(matches)} matches")

            for item in matches:
                if not first_item:
                    first_item = item
                    queue = plex.createPlayQueue(first_item)
                    logger.info(f"Created queue with first item: {item.title}")
                else:
                    queue.addItem(item)
                    logger.info(f"Added to queue: {item.title}")
                queued_items.append(item.title)

        if not first_item:
            logger.error("No matching media found")
            return jsonify({'error': 'No matching media found'}), 404

        client = plex.client("Ryzen-Win")
        client.playMedia(queue)
        logger.info(f"Playing queue with {len(queued_items)} items")

        response = {
            'queue_length': len(queued_items),
            'items': queued_items
        }
        logger.info(f"Returning response: {response}")
        return jsonify(response)

    except Exception as e:
        logger.error(f"Error in queue_specific: {str(e)}", exc_info=True)
        return jsonify({'error': str(e)}), 500


@app.route("/smart-search", methods=['GET'])
def smart_search():
    plex = login()
    term = request.args.get('term', '')
    media_type = request.args.get('type', 'movies')

    library = plex.library.section('Movies' if media_type == 'movies' else 'TV Shows')
    matches = library.search(title=term)

    results = [{
        'title': item.title,
        'year': item.year,
        'summary': item.summary,
        'type': media_type
    } for item in matches]

    return jsonify({
        'results': results,
        'count': len(results)
    })

@app.route("/add-to-queue", methods=['POST'])
def add_to_queue():
    plex = login()
    data = request.json
    titles = data.get('titles', [])
    media_type = data.get('type', 'movie')

    if not titles:
        return jsonify({'error': 'No titles provided'}), 400

    section = plex.library.section('Movies' if media_type == 'movie' else 'TV Shows')

    # Get current queue or create new one
    try:
        client = plex.client("Ryzen-Win")
        queue = client.timeline.playQueue
    except:
        first_item = section.get(titles[0])
        queue = plex.createPlayQueue(first_item)
        titles = titles[1:]

    # Add items to queue
    for title in titles:
        item = section.get(title)
        queue.addItem(item)

    return jsonify({
        'queue_length': len(queue.items),
        'items': [item.title for item in queue.items]
    })


@app.route("/list-media", methods=['GET'])
def list_media():
    try:
        plex = login()
        media_type = request.args.get('type', 'movies')
        page = int(request.args.get('page', 1))
        per_page = int(request.args.get('per_page', 20))
        search_term = request.args.get('search', '')

        logger.info(f"List media request - type: {media_type}, page: {page}, search: {search_term}")

        library = plex.library.section('Movies' if media_type == 'movies' else 'TV Shows')
        all_media = library.search(title=search_term) if search_term else library.all()

        start_idx = (page - 1) * per_page
        end_idx = start_idx + per_page
        page_items = all_media[start_idx:end_idx]

        items = []
        for item in page_items:
            media_item = {
                "title": item.title,
                "year": item.year,
                "summary": item.summary,
                "rating": item.rating,
                "duration": item.duration
            }
            if media_type == 'shows':
                media_item.update({
                    "episode_count": len(item.episodes()),
                    "season_count": len(item.seasons())
                })
            items.append(media_item)

        response = {
            'items': items,
            'total': len(all_media),
            'page': page,
            'total_pages': (len(all_media) + per_page - 1) // per_page
        }
        logger.info(f"Returning {len(items)} items")
        return jsonify(response)

    except Exception as e:
        logger.error(f"Error in list_media: {str(e)}", exc_info=True)
        return jsonify({'error': str(e)}), 500


@app.route("/player-controls", methods=['POST'])
def player_controls():
    try:
        data = request.json
        action = data.get('action')
        logger.info(f"Player control request: {action}")

        plex = login()
        client = plex.client("Ryzen-Win")

        actions = {
            'play': client.play,
            'pause': client.pause,
            'stop': client.stop,
            'skipNext': client.skipNext,
            'skipPrevious': client.skipPrevious,
            'seekTo': lambda: client.seekTo(int(data.get('time', 0))),
            'setVolume': lambda: client.setVolume(int(data.get('volume', 100))),
            'toggleMute': lambda: client.setVolume(0) if client.volume > 0 else client.setVolume(100)
        }

        if action in actions:
            actions[action]()
            logger.info(f"Action {action} executed successfully")
            return jsonify({'status': 'success'})

        logger.error(f"Invalid action: {action}")
        return jsonify({'error': 'Invalid action'}), 400

    except Exception as e:
        logger.error(f"Error in player_controls: {str(e)}", exc_info=True)
        return jsonify({'error': str(e)}), 500

@app.route("/player-status")
def player_status():
    plex = login()
    client = plex.client("Ryzen-Win")
    try:
        timeline = client.timeline
        return jsonify({
            'state': timeline.state if hasattr(timeline, 'state') else 'unknown',
            'time': timeline.time if hasattr(timeline, 'time') else 0,
            'duration': timeline.duration if hasattr(timeline, 'duration') else 0,
            'volume': client.volume if hasattr(client, 'volume') else 100,
            'muted': client.volume == 0 if hasattr(client, 'volume') else False,
            'shuffle': client.shuffled if hasattr(client, 'shuffled') else False,
            'repeat': client.repeat if hasattr(client, 'repeat') else False,
            'current_media': {
                'title': timeline.title if hasattr(timeline, 'title') else None,
                'type': timeline.type if hasattr(timeline, 'type') else None,
                'audio_streams': client.audioStreams if hasattr(client, 'audioStreams') else [],
                'subtitle_streams': client.subtitleStreams if hasattr(client, 'subtitleStreams') else []
            } if hasattr(timeline, 'title') else None
        })
    except Exception as e:
        return jsonify({'error': str(e)}), 404


@app.route("/get-clients")
def get_clients():
    plex = login()
    return jsonify({
        'clients': [
            {
                'name': client.title,
                'device': client.device,
                'model': client.model,
                'platform': client.platform,
                'state': client.state,
                'version': client.version
            }
            for client in plex.clients()
        ]
    })


@app.route("/get-playlists")
def get_playlists():
    plex = login()
    return jsonify({
        'playlists': [
            {
                'title': playlist.title,
                'items': len(playlist.items()),
                'duration': playlist.duration,
                'type': playlist.playlistType
            }
            for playlist in plex.playlists()
        ]
    })

if __name__ == "__main__":
    logger.info(f"Starting {config.ENV} server on {config.HOST}:{config.PORT}")
    logger.info(f"CORS origins: {config.CORS_ORIGINS}")

    if config.ENV == 'development':
        # Use Flask's development server
        app.run(host=config.HOST, port=config.PORT, debug=config.DEBUG)
    else:
        # Use waitress for production
        serve(app, host=config.HOST, port=config.PORT)