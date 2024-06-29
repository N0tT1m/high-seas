from plexapi.myplex import PlexServer
from plexapi.client import PlexClient
from flask import Flask
import random
from flask import request
import config
from flask import Flask
from flask_cors import CORS

app = Flask(__name__)
CORS(app) # This will enable CORS for all routes

def login():
    baseurl = 'http://' + config.IP + ':' + config.PORT
    token = '-yQYUqbbAqgBgKpgsPAm'
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
    for video in movies.search(unwatched=False):
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

    response = flask.jsonify({'movies': movies_list})
    response.headers.add('Access-Control-Allow-Origin', '*')
    return response

@app.route("/shows", methods=['GET'])
def get_shows():
    shows_list = []

    plex = login()
    shows = plex.library.section('TV Shows')
    for show in shows.search(unwatched=False):
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

    response = flask.jsonify({'shows': shows_list})
    response.headers.add('Access-Control-Allow-Origin', '*')
    return response

# ... (omitted anime routes for brevity)

# @app.route("/get-random-movie")
def get_random_movie():
    # num = request.form['number']
    num = "4"

    plex = login()

    devices()

    device_id = 13

    movies = plex.library.section('Movies').all()
    movie_type = random.choice(movies)
    queue = plex.createPlayQueue(movie_type)
    print("Num: ", num)
    for item in range(int(num)):
        movies = plex.library.section('Movies').all()
        movie = random.choice(movies)
        queue.addItem(movie)
        print(queue.items)

    device = plex.systemDevice(int(device_id))
    # client = PlexClient(plex, baseurl="http://192.168.1.66:32400/", token="-yQYUqbbAqgBgKpgsPAm", identifier="di8mvfiy9cmfk91hfaxpnc65")
    # client.playMedia(queue)
