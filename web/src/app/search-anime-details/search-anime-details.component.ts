import { Component, Input, inject } from '@angular/core';
import { CommonModule } from '@angular/common';
import { RouterModule, ActivatedRoute } from '@angular/router';
import { TvShow, TvShowResult } from '../http-service/http-service.component';
import { FormControl, FormGroup, FormsModule, ReactiveFormsModule, NgModel } from '@angular/forms';
import { TvShowService } from '../tv-service.service';

@Component({
  selector: 'app-search-anime-details',
  standalone: true,
  imports: [
    CommonModule,
    RouterModule,
    ReactiveFormsModule,
    FormsModule,
  ],
  providers: [TvShowService, NgModel],
  template: `
  <article class="anime-details" *ngFor="let anime of this.fetchedAnime?.results; index as i;">
    <div class="anime-header">
      <img class="anime-poster" [src]="anime!.poster_path" alt="Poster of {{anime!.name}}"/>
      <div class="anime-info">
        <h2 class="anime-title">{{anime!.name}}</h2>
        <p class="anime-original-title" *ngIf="anime!.original_name !== anime!.name">{{anime!.original_name}}</p>
        <p class="anime-overview">{{anime!.overview}}</p>
      </div>
    </div>
    <section class="anime-details-section">
      <h3 class="section-heading">About this anime</h3>
      <div class="anime-meta">
        <div class="anime-meta-item">
          <span class="anime-meta-label">Original Language:</span>
          <span class="anime-meta-value">{{anime!.original_language}}</span>
        </div>
        <div class="anime-meta-item">
          <span class="anime-meta-label">Popularity:</span>
          <span class="anime-meta-value">{{anime!.popularity}}</span>
        </div>
        <div class="anime-meta-item">
          <span class="anime-meta-label">First Air Date:</span>
          <span class="anime-meta-value">{{anime!.first_air_date}}</span>
        </div>
        <div class="anime-meta-item">
          <span class="anime-meta-label">Number of Seasons:</span>
          <span class="anime-meta-value">{{this.seasonEpisodeNumbers.length}}</span>
        </div>
        <div class="anime-meta-item">
          <span class="anime-meta-label">Number of Episodes:</span>
          <span class="anime-meta-value">{{this.episodes}}</span>
        </div>
        <div class="anime-meta-item">
          <span class="anime-meta-label">Status:</span>
          <span class="anime-meta-value">{{this.status}}</span>
        </div>
      </div>
      <div class="anime-creators anime-meta-item" *ngIf="this.createdBy.length > 0">
        <h4 class="anime-creators-heading">Created By:</h4>
        <ul class="anime-creators-list">
          <li class="anime-creator" *ngFor="let createdBy of this.createdBy; index as j;">{{createdBy['name']}}</li>
        </ul>
      </div>
      <div class="anime-dates">
        <div class="anime-date-item">
          <span class="anime-date-label">First Air Date:</span>
          <span class="anime-date-value">{{this.firstAirDate}}</span>
        </div>
        <div class="anime-date-item">
          <span class="anime-date-label">Last Air Date:</span>
          <span class="anime-date-value">{{this.lastAirDate}}</span>
        </div>
      </div>
      <div class="anime-last-episode" *ngIf="lastEpisodeToAir && lastEpisodeToAir['still_path']">
        <h4 class="anime-last-episode-heading">Last Episode to Air:</h4>
        <div class="anime-last-episode-details">
          <img class="anime-last-episode-image" src="{{this.baseUrl + this.lastEpisodeToAir['still_path']}}" alt="Still from last episode"/>
          <div class="anime-last-episode-info">
            <div class="anime-last-episode-item">
              <span class="anime-last-episode-label">Name:</span>
              <span class="anime-last-episode-value">{{this.lastEpisodeToAir['name']}}</span>
            </div>
            <div class="anime-last-episode-item">
              <span class="anime-last-episode-label">Overview:</span>
              <span class="anime-last-episode-value">{{this.lastEpisodeToAir['overview']}}</span>
            </div>
            <div class="anime-last-episode-item">
              <span class="anime-last-episode-label">Air Date:</span>
              <span class="anime-last-episode-value">{{this.lastEpisodeToAir['air_date']}}</span>
            </div>
            <div class="anime-last-episode-item">
              <span class="anime-last-episode-label">Season:</span>
              <span class="anime-last-episode-value">{{this.lastEpisodeToAir['season_number']}}</span>
            </div>
          </div>
        </div>
      </div>
      <div class="anime-homepage" *ngIf="this.homepage">
        <span class="anime-homepage-label">Homepage:</span>
        <a class="anime-homepage-link" href="{{this.homepage}}" target="_blank">{{this.homepage}}</a>
      </div>
      <div class="anime-production" *ngIf="this.inProduction !== 'false'">
        <span class="anime-production-label">Is this Anime in Production:</span>
        <span class="anime-production-value">Yes</span>
      </div>
      <div class="anime-tagline" *ngIf="this.tagline">
        <span class="anime-tagline-label">Tagline for {{anime.name}}:</span>
        <span class="anime-tagline-value">{{this.tagline}}</span>
      </div>
    </section>
    <div class="anime-meta-item" *ngIf="this.in_plex">
      <div class="anime-meta-label">Status:</div>
      <div class="anime-meta-value">
        <span class="plex-badge">Available in Plex</span>
      </div>
    </div>
    <div class="anime-actions">
      <div class="anime-download-quality">
        <label for="quality" class="anime-download-quality-label">Download Quality:</label>
        <select [(ngModel)]="quality" name="quality" id="quality" class="anime-download-quality-select">
          <option value="4k">4k</option>
          <option value="2k">2k</option>
          <option value="1080p" selected>1080p</option>
          <option value="720p">720p</option>
          <option value="480p">480p</option>
          <option value="240p">240p</option>
        </select>
      </div>
      <button class="anime-download-button" (click)="downloadAnime(anime.name)">Download Anime</button>
    </div>
  </article>
  `,
  styleUrls: ['./search-anime-details.component.sass']
})
export class SearchAnimeDetailsComponent {
  public baseUrl = 'https://image.tmdb.org/t/p/w300_and_h450_bestv2/';
  public route: ActivatedRoute = inject(ActivatedRoute);
  public tvShowService = inject(TvShowService);
  public fetchedData: TvShow[] = [{
    page: 0,
    results: [{
      adult: false,
      backdrop_path: "",
      genre_ids: [],
      id: 0,
      name: "",
      first_air_date: "",
      original_language: "",
      original_name: "",
      overview: "",
      popularity: 0,
      poster_path: "",
      vote_average: 0,
      vote_count: 0,
      video: false
    }],
    total_pages: 0,
    total_result: 0
  }];

  public fetchedAnime: TvShow | undefined;
  public animeList: TvShow[] = [{
    page: 0,
    results: [{
      adult: false,
      backdrop_path: "",
      genre_ids: [],
      id: 0,
      name: "",
      first_air_date: "",
      original_language: "",
      original_name: "",
      overview: "",
      popularity: 0,
      poster_path: "",
      vote_average: 0,
      vote_count: 0,
      video: false
    }],
    total_pages: 0,
    total_result: 0
  }];

  public seasonEpisodeNumbers = [0];
  public totalSeason = [0];
  public status = "";
  public episodes = 0;
  public createdBy = [{}];
  public firstAirDate = "";
  public homepage = "";
  public inProduction = "";
  public lastAirDate = "";
  public lastEpisodeToAir = {};
  public tagline = "";
  public quality = '1080p'; // Default download quality
  public tmdbId: number = 0;
  public overview: string = "";
  public in_plex: boolean = false;

  constructor() {}

  ngOnInit() {
    const animeId = parseInt(this.route.snapshot.params['id'], 10);
    const page = parseInt(this.route.snapshot.params['page'], 10);
    const query = this.route.snapshot.params['query'];

    // Get anime data using discover API with language filter for Japanese shows
    const animeFilters = {
      language: 'ja',
      page: page
    };

    if (query) {
      // Search for anime with query
      this.tvShowService.searchShows(query, animeFilters).subscribe((resp) => {
        this.processAnimeResults(resp, animeId);
      });
    } else {
      // Get anime by discover
      this.tvShowService.discoverShows(animeFilters).subscribe((resp) => {
        this.processAnimeResults(resp, animeId);
      });
    }

    // Get detailed anime information
    this.tvShowService.getTvShowDetails(animeId).subscribe(show => {
      this.processAnimeDetails(show);
    });
  }

  private processAnimeResults(resp: TvShow, animeId: number) {
    if (resp && resp.results) {
      resp.results.forEach((anime) => {
        let result: TvShowResult[] = [{
          adult: anime.adult,
          backdrop_path: anime.backdrop_path,
          genre_ids: anime.genre_ids,
          id: anime.id,
          name: anime.name,
          first_air_date: anime.first_air_date,
          original_language: anime.original_language,
          original_name: anime.original_name,
          overview: anime.overview,
          popularity: anime.popularity,
          poster_path: this.baseUrl + anime.poster_path,
          vote_average: anime.vote_average,
          vote_count: anime.vote_count,
          video: anime.video,
        }];

        this.animeList.push({
          page: resp.page,
          results: result,
          total_pages: resp.total_pages,
          total_result: resp.total_result
        });
      });

      this.animeList.splice(0, 1);

      for (let i = 0; i < this.animeList.length; i++) {
        if (this.animeList[i].results[0] && this.animeList[i].results[0].id === animeId) {
          this.fetchedAnime = this.animeList[i];
          break;
        }
      }
    }
  }

  private processAnimeDetails(show: any) {
    if (show.seasons) {
      this.seasonEpisodeNumbers = [];
      this.totalSeason = [];

      show.seasons.forEach(season => {
        if (season.episode_count > 0) {
          this.seasonEpisodeNumbers.push(season.episode_count);
          this.totalSeason.push(season.season_number);
        }
      });

      this.status = show.status;
      this.createdBy = show.created_by ? show.created_by.filter(creator => creator && Object.keys(creator).length > 0) : [];
      this.firstAirDate = show.first_air_date;
      this.homepage = show.homepage;
      this.inProduction = show.in_production;
      this.lastAirDate = show.last_air_date;
      this.lastEpisodeToAir = show.last_episode_to_air || {};
      this.tagline = show.tagline;
      this.tmdbId = show.id;
      this.overview = show.overview;
      this.in_plex = show.in_plex;

      // Calculate total episodes
      this.episodes = this.seasonEpisodeNumbers.reduce((total, episodes) => total + episodes, 0);
    }
  }

  downloadAnime(title: string) {
    // For anime, always use the anime download request method
    this.tvShowService.makeAnimeShowDownloadRequest(
      title,
      this.seasonEpisodeNumbers,
      this.quality,
      this.tmdbId,
      this.overview
    ).subscribe(
      request => {
        console.log(request);
        // Show the pop-up when the request is successful
        alert('Anime download request submitted successfully!');
      },
      error => {
        console.error(error);
        // Show an error message if the request fails
        alert('An error occurred while submitting the anime download request.');
      }
    );
  }
}
