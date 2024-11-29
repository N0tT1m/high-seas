import { Component, inject, Input } from '@angular/core';
import { CommonModule } from '@angular/common';
import { RouterModule, ActivatedRoute } from '@angular/router';
import { TvShow, TvShowResult } from '../http-service/http-service.component';
import { FormControl, FormGroup, ReactiveFormsModule, FormsModule, NgModel } from '@angular/forms';
import { TvShowService } from '../tv-service.service';

@Component({
  selector: 'app-now-playing-details',
  standalone: true,
  imports: [CommonModule, RouterModule, ReactiveFormsModule, FormsModule],
  providers: [TvShowService, NgModel],
  template: `
    <article class="show-details" *ngFor="let show of this.fetchedShow?.results; index as i;">
      <div class="show-header">
        <img *ngIf="show!.poster_path" class="show-poster" [src]="show!.poster_path" alt="Poster of {{show!.name}}" />
        <div class="show-info">
          <h2 class="show-title">{{show!.name}}</h2>
          <p class="show-overview">{{show!.overview}}</p>
        </div>
      </div>
      <section class="show-details-section">
        <h3 class="section-heading">About this show:</h3>
        <div class="show-meta">
          <div class="show-meta-item">
            <div class="show-meta-label">Original Language:</div>
            <div class="show-meta-value">{{show!.original_language}}</div>
          </div>
          <div class="show-meta-item">
            <div class="show-meta-label">Original Title:</div>
            <div class="show-meta-value">{{show!.original_name}}</div>
          </div>
          <div class="show-meta-item">
            <div class="show-meta-label">Popularity:</div>
            <div class="show-meta-value">{{show!.popularity}}</div>
          </div>
          <div class="show-meta-item">
            <div class="show-meta-label">First Air Date:</div>
            <div class="show-meta-value">{{show!.first_air_date}}</div>
          </div>
          <div class="show-meta-item">
            <div class="show-meta-label">Number of Seasons:</div>
            <div class="show-meta-value">{{this.seasonEpisodeNumbers.length}}</div>
          </div>
          <div class="show-meta-item">
            <div class="show-meta-label">Number of Episodes:</div>
            <div class="show-meta-value">{{this.episodes}}</div>
          </div>
          <div class="show-meta-item">
            <div class="show-meta-label">Status of this Show:</div>
            <div class="show-meta-value">{{this.status}}</div>
          </div>
          <div class="show-creators show-meta-item" *ngIf="this.createdBy.length > 0">
            <h4 class="show-creators-heading">Created By:</h4>
            <ul class="show-creators-list">
              <li class="show-creator" *ngFor="let createdBy of this.createdBy; index as j;">{{createdBy['name']}}</li>
            </ul>
          </div>
        </div>
        <div class="show-last-episode" *ngIf="this.lastEpisodeToAir['still_path']">
          <h4 class="show-last-episode-heading">Last Episode to Air:</h4>
          <div class="show-last-episode-details">
            <img class="show-last-episode-image" src="{{this.baseUrl + this.lastEpisodeToAir['still_path']}}" alt="Still from last episode" />
          </div>
          <div class="show-last-episode-info">
              <div class="show-last-episode-item">
                <span class="show-last-episode-label">Name:</span>
                <span class="show-last-episode-value">{{this.lastEpisodeToAir['name']}}</span>
              </div>
              <div class="show-last-episode-item">
                <span class="show-last-episode-label">Overview:</span>
                <span class="show-last-episode-value">{{this.lastEpisodeToAir['overview']}}</span>
              </div>
              <div class="show-last-episode-item">
                <span class="show-last-episode-label">Air Date:</span>
                <span class="show-last-episode-value">{{this.lastEpisodeToAir['air_date']}}</span>
              </div>
              <div class="show-last-episode-item">
                <span class="show-last-episode-label">Season:</span>
                <span class="show-last-episode-value">{{this.lastEpisodeToAir['season_number']}}</span>
              </div>
            </div>
        </div>
        <div class="show-homepage">
          <span class="show-homepage-label">Homepage:</span>
          <a class="show-homepage-link" href="{{this.homepage}}" target="_blank">{{this.homepage}}</a>
        </div>
        <div class="show-production" *ngIf="this.inProduction != 'false'">
          <span class="show-production-label">Is this Show in Production:</span>
          <span class="show-production-value">Yes</span>
        </div>
        <div class="show-tagline">
          <span class="show-tagline-label">Tagline for {{show.name}}:</span>
          <span class="show-tagline-value">{{this.tageline}}</span>
        </div>
        <div class="show-video" *ngIf="show!.video != undefined">
          <span class="show-video-label">Is a video:</span>
          <span class="show-video-value">{{show!.video}}</span>
        </div>
      </section>
      <div class="show-actions">
        <div class="show-download-quality">
          <label for="quality" class="show-download-quality-label">Download Quality:</label>
          <select [(ngModel)]="quality" name="quality" id="quality" class="show-download-quality-select">
            <option value="4k">4k</option>
            <option value="2k">2k</option>
            <option value="1080p" selected>1080p</option>
            <option value="720p">720p</option>
            <option value="480p">480p</option>
            <option value="240p">240p</option>
          </select>
        </div>
        <button class="show-download-button" (click)="downloadShow(show.name, show.original_language, this.quality)">Download Show</button>
      </div>
    </article>
  `,
  styleUrls: ['airing-today-gallery-tv-shows-details.component.sass', '../../styles.sass']
})

export class AiringTodayGalleryTvShowsDetailsComponent {
  public baseUrl = 'https://image.tmdb.org/t/p/w300_and_h450_bestv2/';
  public route: ActivatedRoute = inject(ActivatedRoute);
  public tvShowService = inject(TvShowService);
  public fetchedData: TvShow[] = [{page: 0, results: [{adult: false, backdrop_path: "", genre_ids: [], id: 0, name: "", first_air_date: "", original_language: "", original_name: "", overview: "", popularity: 0, poster_path: "", vote_average: 0, vote_count: 0, video: false}], total_pages: 0, total_result: 0}]
  public fetchedShow: TvShow | undefined;
  public tvShowList: TvShow[] = [{page: 0, results: [{adult: false, backdrop_path: "", genre_ids: [], id: 0, name: "", first_air_date: "", original_language: "", original_name: "", overview: "", popularity: 0, poster_path: "", vote_average: 0, vote_count: 0, video: false}], total_pages: 0, total_result: 0}]
  public showsLength: number;
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
  public tageline = "";
  public quality = '1080p'; // Default download quality
  public tmdbId: number = 0;
  public overview: string = "";

  constructor() {

  }

  ngOnInit() {
    const tvShowId = parseInt(this.route.snapshot.params['id'], 10);
    const page = parseInt(this.route.snapshot.params['page'], 10);
    const query = this.route.snapshot.params['query'];

    this.tvShowService.getAiringToday(page).subscribe((resp) => {
      if (resp && resp['results']) {
        resp['results'].forEach((show) => {
          let page = resp['page'];
          let isAdult = show['adult'];
          let backdropPath = show['backdrop_path'];
          let genreIds = show['genre_ids'];
          let id = show['id'];
          let firstAirDate = show['first_air_date'];
          let video = show['video'];
          let name = show['name'];
          let originalLanguage = show['original_language'];
          let originalName = show['original_name'];
          let overview = show['overview'];
          let popularity = show['popularity'];
          let posterPath = this.baseUrl + show['poster_path'];
          let voteAverage = show['vote_average'];
          let voteCount = show['vote_count'];
          let totalPages = resp['total_pages'];
          let totalResult = resp['total_results'];

          let result: TvShowResult[] = [{
            adult: isAdult,
            backdrop_path: backdropPath,
            genre_ids: genreIds,
            id: id,
            name: name,
            first_air_date: firstAirDate,
            original_language: originalLanguage,
            original_name: originalName,
            overview: overview,
            popularity: popularity,
            poster_path: posterPath,
            vote_average: voteAverage,
            vote_count: voteCount,
            video: video
          }];

          this.tvShowList.push({
            page: page,
            results: result,
            total_pages: totalPages,
            total_result: totalResult
          });
        });

        this.tvShowList.splice(0, 1);

        for (var i = 0; i < this.tvShowList.length; i++) {
          this.fetchedShow = this.tvShowList.find(movieResult => movieResult.results[i]!.id === tvShowId);
        }
      } else {
        console.error('Invalid response format. "results" property not found.');
      }
    });

    this.tvShowService.getShowDetails(tvShowId).subscribe(show => {
      this.showsLength = show.seasons.length;
      for (var i = 0; i < this.showsLength; i++) {
        this.seasonEpisodeNumbers.push(show.seasons[i]['episode_count'])
        this.totalSeason.push(show.seasons[i]['season_number'])
        this.status = show.status
        if (show.created_by[i] && Object.keys(show.created_by[i]).length > 0) {
          this.createdBy.push(show.created_by[i])
        }
        this.firstAirDate = show.first_air_date;
        this.homepage = show.homepage;
        this.inProduction = show.in_production;
        this.lastAirDate = show.last_air_date;
        this.lastEpisodeToAir = show.last_episode_to_air;
        this.tageline = show.tagline;
        this.tmdbId = show.id;
        this.overview = show.overview;
      }

      this.seasonEpisodeNumbers.splice(0, 1);
      this.totalSeason.splice(0, 1);
      this.createdBy.splice(0, 1);

      for (let i = 0; i < this.seasonEpisodeNumbers.length; i++) {
        this.episodes = this.episodes + this.seasonEpisodeNumbers[i];
      }
    })
  }


  downloadShow(title: string, lang: string, quality: string) {
    if (lang === 'ja') {
      console.log('ANIME');
      this.tvShowService.makeAnimeDownloadRequest(title, this.seasonEpisodeNumbers, this.quality, this.tmdbId, this.overview).subscribe(request => console.log(request))
    } else {
      console.log('Movie');
      this.tvShowService.makeTvShowDownloadRequest(title, this.seasonEpisodeNumbers, this.quality, this.tmdbId, this.overview).subscribe(
        request => {
          console.log(request);
          // Show the pop-up when the request is successful
          alert('Download request submitted successfully!');
        },
        error => {
          console.error(error);
          // Show an error message if the request fails
          alert('An error occurred while submitting the download request.');
        }
      );
    }
  }
}
