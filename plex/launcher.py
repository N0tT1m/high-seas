from plexapi.myplex import PlexServer
from flask import Flask
import random
from flask import request
import config

app = Flask(__name__)

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

@app.route("/movies")
def get_movies():
    movies_list = []
    
    plex = login()
    movies = plex.library.section('Movies')
    for video in movies.search(unwatched=False):
        movies_list.append({"title": video.title})
        
    return movies_list
        
@app.route("/shows")
def get_shows():
    shows_list = []
    
    plex = login()
    movies = plex.library.section('TV Shows')
    for video in shows.search(unwatched=False):
        shows_list.append({"name": video.title})
    
    return shows_list
       
@app.route("/anime-subbed-movies") 
def get_anime_subbed_movies():
    movies_list = []
    
    plex = login()
    movies = plex.library.section('Anime Subbed Movies')
    for video in movies.search(unwatched=False):
        movies_list.append({"title": video.title})
        
    return movies_list

@app.route("/animed-dubbed-movies")
def get_anime_dubbed_movies():
    movies_list = []
    
    plex = login()
    movies = plex.library.section('Anime Dubbed Movies')
    for video in movies.search(unwatched=False):
        movies_list.append({"title": video.title})
        
    return movies_list
        
@app.route("/anime-subbed-shows")
def get_anime_subbed_shows():
    shows_list = []
    
    plex = login()
    movies = plex.library.section('Anime Subbed Shows')
    for video in shows.search(unwatched=False):
        shows_list.append({"name": video.title})
    
    return shows_list

@app.route("/anime-dubbed-shows")
def get_anime_dubbed_shows():
    shows_list = []
    
    plex = login()
    movies = plex.library.section('Anime Dubbed Shows')
    for video in shows.search(unwatched=False):
        shows_list.append({"name": video.title})
    
    return shows_list

# @app.route("/get-random-movie")
def get_random_movie():
    # num = request.form['number']
    num = "4"
    
    plex = login()
    
    devices()
    
    device_id = 4
    
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

    client = plex.client("Pixel-6-Pro")
    client.playMedia(queue)

    
get_random_movie()

# @app.route("/get-random-movie")
# def get_random_show():
    
    
# @app.route("/get-random-movie")
# def get_random_anime_movie():
    

# @app.route("/get-random-movie")
# def get_random_anime_show():  

