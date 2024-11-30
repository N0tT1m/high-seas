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
import requests

PLEX_TIMEOUT = 60

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
    """Initialize Plex server connection with proper timeout and retry handling"""
    baseurl = 'http://' + config.PLEX_IP + ':' + config.PLEX_PORT
    token = 'Y7fU6x3PPqr8A-P3WEjq'

    try:
        plex = PlexServer(baseurl, token, timeout=PLEX_TIMEOUT)
        return plex
    except Exception as e:
        logger.error(f"Failed to connect to Plex server: {str(e)}")
        raise

# Error handler for 404
@app.errorhandler(404)
def not_found(e):
    return jsonify({
        'error': 'Endpoint not found',
        'status': 'error'
    }), 404

# Error handler for 500
@app.errorhandler(500)
def server_error(e):
    return jsonify({
        'error': str(e),
        'status': 'error'
    }), 500


@app.route("/player-controls", methods=['POST'])
def player_controls():
    try:
        data = request.json
        action = data.get('action')
        logger.info(f"Player control request: {action}")

        plex = login()
        try:
            client = plex.client("Ryzen-Win")
        except Exception as e:
            logger.error(f"Could not connect to client: {str(e)}")
            return jsonify({
                'error': 'Could not connect to Plex client',
                'status': 'error'
            }), 500

        try:
            if action == 'play':
                client.play()
            elif action == 'pause':
                client.pause()
            elif action == 'stop':
                client.stop()
            elif action == 'skipNext':
                client.skipNext()
            elif action == 'skipPrevious':
                client.skipPrevious()
            elif action == 'seekTo':
                time_ms = int(data.get('time', 0))
                client.seekTo(time_ms)
            elif action == 'setVolume':
                volume = int(data.get('volume', 100))
                client.setVolume(volume)
            elif action == 'mute':
                client.setVolume(0)
            elif action == 'unmute':
                client.setVolume(100)
            else:
                logger.error(f"Invalid action: {action}")
                return jsonify({
                    'error': 'Invalid action',
                    'status': 'error'
                }), 400

            logger.info(f"Successfully executed action: {action}")
            return jsonify({
                'status': 'success',
                'message': f'Successfully executed {action}'
            })

        except Exception as e:
            logger.error(f"Error executing player action: {str(e)}")
            return jsonify({
                'error': f'Failed to execute {action}: {str(e)}',
                'status': 'error'
            }), 500

    except Exception as e:
        logger.error(f"Error in player_controls: {str(e)}", exc_info=True)
        return jsonify({
            'error': str(e),
            'status': 'error'
        }), 500


@app.route("/add-to-queue", methods=['POST'])
def add_to_queue():
    try:
        data = request.json
        titles = data.get('titles', [])
        media_type = data.get('type', 'movie')

        logger.info(f"Add to queue request - type: {media_type}, titles: {titles}")

        if not titles:
            return jsonify({
                'error': 'No titles provided',
                'status': 'error'
            }), 400

        plex = login()
        section = plex.library.section('Movies' if media_type == 'movie' else 'TV Shows')
        queued_items = []

        # Try to get current queue
        try:
            queue = plex.playQueue()
            if not queue:
                # If no queue exists, create new one with first item
                first_item = section.search(title=titles[0])[0]
                queue = plex.createPlayQueue(first_item)
                queued_items.append(first_item.title)
                titles = titles[1:]
            logger.info("Using existing queue or created new one")
        except Exception as e:
            logger.warning(f"Could not get/create queue: {str(e)}")
            return jsonify({
                'error': 'Failed to access queue',
                'status': 'error'
            }), 500

        # Add remaining items to queue
        for title in titles:
            try:
                matches = section.search(title=title)
                if matches:
                    item = matches[0]
                    queue.addItem(item)
                    queued_items.append(item.title)
                    logger.info(f"Added {item.title} to queue")
                else:
                    logger.warning(f"No match found for: {title}")
            except Exception as e:
                logger.error(f"Error adding {title} to queue: {str(e)}")
                continue

        if not queued_items:
            return jsonify({
                'error': 'No items were queued',
                'status': 'error',
                'queue_length': 0,
                'items': []
            }), 404

        response = {
            'status': 'success',
            'queue_length': len(queued_items),
            'items': queued_items
        }

        logger.info(f"Successfully added items to queue: {response}")
        return jsonify(response)

    except Exception as e:
        logger.error(f"Error in add_to_queue: {str(e)}", exc_info=True)
        return jsonify({
            'error': str(e),
            'status': 'error',
            'queue_length': 0,
            'items': []
        }), 500


@app.route("/devices", methods=['GET'])
def get_devices():
    try:
        logger.info("Getting system devices")
        plex = login()
        devices = []

        for device in plex.systemDevices():
            devices.append({
                'name': device.name,
                'product': device.product,
                'platform': device.platform,
                'state': device.state,
                'version': device.version,
                'device': device.device,
                'model': device.model,
                'connections': device.connections
            })

        logger.info(f"Found {len(devices)} devices")
        return jsonify({
            'status': 'success',
            'devices': devices
        })

    except Exception as e:
        logger.error(f"Error getting devices: {str(e)}", exc_info=True)
        return jsonify({
            'error': str(e),
            'status': 'error',
            'devices': []
        }), 500


@app.route("/timeline", methods=['GET'])
def get_timeline():
    try:
        logger.info("Getting playback timeline")
        plex = login()
        client = plex.client("Ryzen-Win")
        timeline = client.timeline

        response = {
            'status': 'success',
            'state': timeline.state if hasattr(timeline, 'state') else 'stopped',
            'time': timeline.time if hasattr(timeline, 'time') else 0,
            'duration': timeline.duration if hasattr(timeline, 'duration') else 0,
            'type': timeline.type if hasattr(timeline, 'type') else None,
            'current_title': timeline.title if hasattr(timeline, 'title') else None,
            'shuffle': client.shuffled if hasattr(client, 'shuffled') else False,
            'repeat': client.repeat if hasattr(client, 'repeat') else False,
            'volume': client.volume if hasattr(client, 'volume') else 100,
            'muted': client.volume == 0 if hasattr(client, 'volume') else False,
        }

        logger.info(f"Timeline status: {response}")
        return jsonify(response)

    except Exception as e:
        logger.error(f"Error getting timeline: {str(e)}", exc_info=True)
        return jsonify({
            'error': str(e),
            'status': 'error'
        }), 500

def get_current_queue(plex):
    """Get the current play queue if it exists"""
    try:
        return plex.playqueue()
    except Exception as e:
        logger.warning(f"No existing play queue found: {str(e)}")
        return None

def get_or_create_queue(plex, first_item=None):
    """Get existing queue or create new one with optional first item"""
    try:
        queue = get_current_queue(plex)
        if queue and queue.items:
            logger.info("Using existing play queue")
            return queue
    except Exception as e:
        logger.warning(f"Error getting current queue: {str(e)}")

    if first_item:
        logger.info(f"Creating new queue with first item: {first_item.title}")
        return plex.createPlayQueue(first_item)
    return None


@app.route("/movies", methods=['GET'])
def get_movies():
    try:
        logger.info("Getting unwatched movies count")
        plex = login()
        movies = plex.library.section('Movies')
        unwatched = movies.search(unwatched=True)

        response = {
            'status': 'success',
            'movies': len(unwatched)
        }

        logger.info(f"Found {len(unwatched)} unwatched movies")
        return jsonify(response)
    except Exception as e:
        logger.error(f"Error getting movies: {str(e)}", exc_info=True)
        return jsonify({
            'status': 'error',
            'error': str(e),
            'movies': 0
        }), 500


@app.route("/shows", methods=['GET'])
def get_shows():
    try:
        logger.info("Getting unwatched shows count")
        plex = login()
        shows = plex.library.section('TV Shows')
        unwatched = shows.search(unwatched=True)

        response = {
            'status': 'success',
            'shows': len(unwatched)
        }

        logger.info(f"Found {len(unwatched)} unwatched shows")
        return jsonify(response)
    except Exception as e:
        logger.error(f"Error getting shows: {str(e)}", exc_info=True)
        return jsonify({
            'status': 'error',
            'error': str(e),
            'shows': 0
        }), 500


@app.route("/get-random-media", methods=['POST'])
def get_random_media():
    try:
        # Log request data
        data = request.get_json()
        logger.info(f"Request data: {data}")

        count = int(data.get('number', 1))
        media_type = data.get('type', 'movie')
        use_existing_queue = data.get('use_existing_queue', False)

        logger.info(f"Parsed parameters - count: {count}, type: {media_type}, use_existing_queue: {use_existing_queue}")

        data = request.json
        count = int(data.get('number', 1))
        media_type = data.get('type', 'movie')
        use_existing_queue = data.get('use_existing_queue', False)

        logger.info(f"Random {media_type} request - count: {count}, use existing queue: {use_existing_queue}")

        try:
            plex = login()
        except Exception as e:
            logger.error(f"Failed to login to Plex: {str(e)}")
            return jsonify({
                'error': 'Failed to connect to Plex server',
                'details': str(e)
            }), 500

        library = plex.library.section('Movies' if media_type == 'movie' else 'TV Shows')

        try:
            available_media = library.search(unwatched=True)
        except Exception as e:
            logger.error(f"Failed to search library: {str(e)}")
            return jsonify({
                'error': 'Failed to search library',
                'details': str(e)
            }), 500

        if not available_media:
            return jsonify({
                'error': 'No unwatched media available',
                'queue_length': 0,
                'items': []
            }), 404

        # Select random items
        selected_items = random.sample(available_media, min(count, len(available_media)))
        queued_items = []

        # Handle queue creation/addition
        try:
            if use_existing_queue:
                try:
                    queue = plex.playQueue()
                    if not queue:
                        queue = plex.createPlayQueue(selected_items[0])
                        queued_items.append(selected_items[0].title)
                        for item in selected_items[1:]:
                            queue.addItem(item)
                            queued_items.append(item.title)
                    else:
                        for item in selected_items:
                            queue.addItem(item)
                            queued_items.append(item.title)
                except:
                    queue = plex.createPlayQueue(selected_items[0])
                    queued_items.append(selected_items[0].title)
                    for item in selected_items[1:]:
                        queue.addItem(item)
                        queued_items.append(item.title)
            else:
                queue = plex.createPlayQueue(selected_items[0])
                queued_items.append(selected_items[0].title)
                for item in selected_items[1:]:
                    queue.addItem(item)
                    queued_items.append(item.title)

            # Try to start playback
            try:
                client = plex.client("Ryzen-Win")
                client.playMedia(queue)
            except Exception as e:
                logger.warning(f"Failed to start playback: {str(e)}")

            response_data = {
                'status': 'success',
                'queue_length': len(queued_items),
                'items': queued_items
            }

            logger.info(f"Sending response: {response_data}")
            return jsonify(response_data)

        except Exception as e:
            error_response = {
                'error': str(e),
                'queue_length': 0,
                'items': []
            }
            logger.error(f"Error in get_random_media: {e}", exc_info=True)
            logger.info(f"Sending error response: {error_response}")
            return jsonify(error_response), 500

    except Exception as e:
            logger.error(f"Error in get_random_media: {str(e)}", exc_info=True)
            return jsonify({
                'error': str(e),
                'queue_length': 0,
                'items': []
            }), 500


@app.route("/queue-specific", methods=['POST'])
def queue_specific():
    try:
        data = request.json
        logger.info(f"Received queue request: {data}")

        media_type = data.get('type', 'movie')
        titles = data.get('items', [])
        use_existing_queue = data.get('use_existing_queue', False)
        get_all = data.get('get_all', False)  # New parameter to control if we want all matches

        if not titles:
            return jsonify({'error': 'No titles provided'}), 400

        plex = login()
        library = plex.library.section('Movies' if media_type == 'movie' else 'TV Shows')

        queued_items = []
        queue = None

        if use_existing_queue:
            try:
                queue = plex.playQueue()
            except Exception as e:
                logger.warning(f"Could not get existing queue: {str(e)}")

        # Process each search term
        for search_term in titles:
            try:
                logger.info(f"Searching for: {search_term}")
                matches = library.search(title=search_term)
                logger.info(f"Found {len(matches)} matches for {search_term}")

                if not matches:
                    logger.warning(f"No matches found for: {search_term}")
                    continue

                # Determine which items to queue
                items_to_queue = matches if get_all else [matches[0]]

                for item in items_to_queue:
                    try:
                        if not queue:
                            queue = plex.createPlayQueue(item)
                            logger.info(f"Created queue with first item: {item.title}")
                        else:
                            queue.addItem(item)
                            logger.info(f"Added to queue: {item.title}")
                        queued_items.append(item.title)
                    except Exception as e:
                        logger.error(f"Error adding {item.title} to queue: {str(e)}")
                        continue

            except Exception as e:
                logger.error(f"Error searching for {search_term}: {str(e)}")
                continue

        if not queue:
            logger.error("No matching media found")
            return jsonify({'error': 'No matching media found'}), 404

        try:
            client = plex.client("Ryzen-Win")
            client.playMedia(queue)
            logger.info(f"Playing queue with {len(queued_items)} items")
        except Exception as e:
            logger.error(f"Error starting playback: {str(e)}")
            return jsonify({
                'status': 'partial_success',
                'warning': 'Queue created but playback failed',
                'queue_length': len(queued_items),
                'items': queued_items
            })

        response = {
            'status': 'success',
            'queue_length': len(queued_items),
            'items': queued_items
        }
        logger.info(f"Returning response: {response}")
        return jsonify(response)

    except Exception as e:
        logger.error(f"Error in queue_specific: {str(e)}", exc_info=True)
        return jsonify({
            'error': str(e),
            'queue_length': 0,
            'items': []
        }), 500


@app.route("/smart-search", methods=['GET'])
def smart_search():
    try:
        term = request.args.get('term', '')
        media_type = request.args.get('type', 'movies')

        logger.info(f"Smart search request - term: {term}, type: {media_type}")

        if not term:
            return jsonify({
                'status': 'error',
                'error': 'No search term provided',
                'results': [],
                'count': 0
            }), 400

        plex = login()
        library = plex.library.section('Movies' if media_type == 'movies' else 'TV Shows')

        matches = library.search(title=term)
        results = []

        for item in matches:
            results.append({
                'title': item.title,
                'year': item.year,
                'summary': item.summary,
                'type': media_type
            })

        response = {
            'status': 'success',
            'results': results,
            'count': len(results)
        }

        logger.info(f"Found {len(results)} matches for '{term}'")
        return jsonify(response)
    except Exception as e:
        logger.error(f"Error in smart search: {str(e)}", exc_info=True)
        return jsonify({
            'status': 'error',
            'error': str(e),
            'results': [],
            'count': 0
        }), 500


@app.route("/player-status", methods=['GET'])
def get_player_status():
    try:
        logger.info("Getting player status")
        plex = login()
        client = plex.client("Ryzen-Win")
        timeline = client.timeline

        status = {
            'state': timeline.state if hasattr(timeline, 'state') else 'stopped',
            'time': timeline.time if hasattr(timeline, 'time') else 0,
            'duration': timeline.duration if hasattr(timeline, 'duration') else 0,
            'volume': client.volume if hasattr(client, 'volume') else 100,
            'muted': client.volume == 0 if hasattr(client, 'volume') else False
        }

        if hasattr(timeline, 'title'):
            status['current_media'] = {
                'title': timeline.title,
                'type': timeline.type if hasattr(timeline, 'type') else None
            }

        logger.info(f"Current player status: {status}")
        return jsonify(status)
    except Exception as e:
        logger.error(f"Error getting player status: {str(e)}", exc_info=True)
        return jsonify({
            'error': str(e),
            'state': 'error'
        }), 500


@app.route("/get-clients", methods=['GET'])
def get_clients():
    try:
        logger.info("Getting available clients")
        plex = login()
        clients = []

        for client in plex.clients():
            clients.append({
                'name': client.title,
                'device': client.device,
                'model': client.model,
                'platform': client.platform,
                'state': client.state,
                'version': client.version
            })

        logger.info(f"Found {len(clients)} clients")
        return jsonify({'clients': clients})
    except Exception as e:
        logger.error(f"Error getting clients: {str(e)}", exc_info=True)
        return jsonify({
            'error': str(e),
            'clients': []
        }), 500


@app.route("/get-playlists", methods=['GET'])
def get_playlists():
    try:
        logger.info("Getting playlists")
        plex = login()
        playlists = []

        for playlist in plex.playlists():
            playlists.append({
                'title': playlist.title,
                'items': len(playlist.items()),
                'duration': playlist.duration,
                'type': playlist.playlistType
            })

        logger.info(f"Found {len(playlists)} playlists")
        return jsonify({'playlists': playlists})
    except Exception as e:
        logger.error(f"Error getting playlists: {str(e)}", exc_info=True)
        return jsonify({
            'error': str(e),
            'playlists': []
        }), 500


@app.route("/get-current-queue", methods=['GET'])
def get_queue_status():
    try:
        logger.info("Getting current queue status")
        plex = login()

        try:
            queue = plex.playQueue()
            client = plex.client("Ryzen-Win")
            timeline = client.timeline

            if not queue or not queue.items:
                return jsonify({
                    'message': 'No active queue found',
                    'queue_length': 0,
                    'items': []
                })

            # Get currently playing item
            current_item = None
            if hasattr(timeline, 'title'):
                current_item = timeline.title

            items = []
            for item in queue.items:
                item_data = {
                    'title': item.title,
                    'type': item.type,
                    'duration': item.duration,
                    'selected': (current_item == item.title if current_item else False)
                }
                items.append(item_data)

            response = {
                'status': 'success',
                'queue_length': len(items),
                'current_item': current_item,
                'items': items
            }

            logger.info(f"Returning queue status: {response}")
            return jsonify(response)

        except Exception as e:
            logger.error(f"Error getting queue: {str(e)}")
            return jsonify({
                'message': 'No active queue found or error getting queue status',
                'queue_length': 0,
                'items': [],
                'error': str(e)
            }), 404

    except Exception as e:
        logger.error(f"Error in get_queue_status: {str(e)}", exc_info=True)
        return jsonify({
            'error': str(e),
            'queue_length': 0,
            'items': []
        }), 500


@app.route("/clear-queue", methods=['POST'])
def clear_queue():
    try:
        logger.info("Attempting to clear queue")
        plex = login()

        try:
            queue = plex.playQueue()
            if not queue:
                return jsonify({
                    'message': 'No active queue to clear',
                    'status': 'success'
                })

            # Stop current playback
            try:
                client = plex.client("Ryzen-Win")
                client.stop()
                logger.info("Stopped current playback")
            except Exception as e:
                logger.warning(f"Could not stop playback: {str(e)}")

            # Clear the queue - create a new empty queue
            try:
                if len(queue.items) > 0:
                    first_item = queue.items[0]
                    new_queue = plex.createPlayQueue(first_item)
                    new_queue.clear()
                    logger.info("Queue cleared successfully")
                    return jsonify({
                        'message': 'Queue cleared successfully',
                        'status': 'success'
                    })
            except Exception as e:
                logger.error(f"Error clearing queue: {str(e)}")
                return jsonify({
                    'error': f'Failed to clear queue: {str(e)}',
                    'status': 'error'
                }), 500

        except Exception as e:
            logger.warning(f"No active queue found: {str(e)}")
            return jsonify({
                'message': 'No active queue to clear',
                'status': 'success'
            })

    except Exception as e:
        logger.error(f"Error in clear_queue: {str(e)}", exc_info=True)
        return jsonify({
            'error': str(e),
            'status': 'error'
        }), 500

@app.route("/list-media", methods=['GET'])
def list_media():
    try:
        media_type = request.args.get('type', 'movies')
        page = int(request.args.get('page', 1))
        per_page = int(request.args.get('per_page', 20))
        search_term = request.args.get('search', '')

        logger.info(f"List media request - type: {media_type}, page: {page}, search: {search_term}")

        try:
            plex = login()
        except Exception as e:
            logger.error(f"Failed to connect to Plex: {str(e)}")
            return jsonify({
                'error': 'Failed to connect to Plex server',
                'items': [],
                'total': 0,
                'page': page,
                'total_pages': 0
            }), 500

        library = plex.library.section('Movies' if media_type == 'movies' else 'TV Shows')

        try:
            # Get all media items matching the search term
            if search_term:
                all_media = library.search(title=search_term)
            else:
                all_media = library.all()

            # Sort by title
            all_media.sort(key=lambda x: x.title.lower())

            # Calculate pagination
            total_items = len(all_media)
            total_pages = (total_items + per_page - 1) // per_page
            page = min(max(1, page), total_pages if total_pages > 0 else 1)
            start_idx = (page - 1) * per_page
            end_idx = min(start_idx + per_page, total_items)

            items = []
            for item in all_media[start_idx:end_idx]:
                media_item = {
                    "title": item.title,
                    "year": item.year,
                    "summary": item.summary,
                    "rating": float(item.rating) if item.rating else 0.0,
                    "duration": item.duration,
                }

                if media_type == 'shows':
                    try:
                        media_item.update({
                            "episode_count": len(item.episodes()),
                            "season_count": len(item.seasons()),
                        })
                    except:
                        media_item.update({
                            "episode_count": 0,
                            "season_count": 0,
                        })

                items.append(media_item)

            response = {
                'items': items,
                'total': total_items,
                'page': page,
                'total_pages': total_pages,
            }

            logger.info(f"Returning {len(items)} items")
            return jsonify(response)

        except Exception as e:
            logger.error(f"Error processing media list: {str(e)}", exc_info=True)
            return jsonify({
                'error': f'Error processing media list: {str(e)}',
                'items': [],
                'total': 0,
                'page': page,
                'total_pages': 0
            }), 500

    except Exception as e:
        logger.error(f"Error in list_media: {str(e)}", exc_info=True)
        return jsonify({
            'error': str(e),
            'items': [],
            'total': 0,
            'page': page,
            'total_pages': 0
        }), 500

if __name__ == "__main__":
    logger.info(f"Starting {config.ENV} server on {config.HOST}:{config.PORT}")
    logger.info(f"CORS origins: {config.CORS_ORIGINS}")

    if config.ENV == 'development':
        # Use Flask's development server
        app.run(host=config.HOST, port=config.PORT, debug=config.DEBUG)
    else:
        # Use waitress for production
        serve(app, host=config.HOST, port=config.PORT)