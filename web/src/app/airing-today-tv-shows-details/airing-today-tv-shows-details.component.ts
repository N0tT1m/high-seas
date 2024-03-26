import { Component, inject, Input } from '@angular/core';
import { CommonModule } from '@angular/common';
import { RouterModule, ActivatedRoute } from '@angular/router';
import { TvShow, TvShowResult } from '../http-service/http-service.component';
import { FormControl, FormGroup, ReactiveFormsModule, FormsModule, NgModel } from '@angular/forms';
import { TvShowService } from '../tv-service.service';

@Component({
  selector: 'app-now-playing-details',
  standalone: true,
  imports: [
    CommonModule,
    RouterModule,
    ReactiveFormsModule,
    FormsModule,
  ],
  providers: [TvShowService, NgModel],
  template: `
  <article class="the-girls" *ngFor="let show of this.fetchedShow?.results; index as i;">
  <img class="show-photo the-girls" [src]="show!.poster_path"
        alt="Exterior photo of {{show!.name}}"/>
      <section class="show-description the-girls">
        <h2 class="show-title">{{show!.name}}</h2>
        <p class="show-overview">{{show!.overview}}</p>
      </section>
      <section class="show the-girls">
        <h3 class="section-heading">About this show</h3>
        <ul>
          <div class="show-div">
            <li class="show-details">Original Language: {{show!.original_language}}</li>
            <li class="show-details">Original Title: {{show!.original_name}}</li>
            <li class="show-details">Popularity: {{show!.popularity}}</li>
            <li class="show-details">First Air Date: {{show!.first_air_date}}</li>
            <li class="show-details">Number of Seasons: {{this.seasonEpisodeNumbers.length}}</li>
            <li class="show-details">Number of Episodes: {{this.episodes}}</li>
            <li class="show-details">Status of this Show: {{this.status}}</li>
          </div>
          <div class="show-div">
            <p class="created-by">Created By:</p>
            <li class="created-by" *ngFor="let createdBy of this.createdBy; index as j;">{{createdBy['name']}}</li>
          </div>
          <div class="show-div">
            <li class="air-date">First Air Date: {{this.firstAirDate}}</li>
            <li class="air-date">Last Air Date: {{this.lastAirDate}}</li>
          </div>
          <div class="show-div">
            <p class="last-episode">Last Episode to Air:</p>
            <img class="last-episode" src="{{this.baseUrl + this.lastEpisodeToAir['still_path']}}" />
            <li class="last-episode">Last Episode Name: {{this.lastEpisodeToAir['name']}}</li>
            <li class="last-episode">Last Episode Overview: {{this.lastEpisodeToAir['overview']}}</li>
            <li class="last-episode">Last Episode Air Date: {{this.lastEpisodeToAir['air_date']}}</li>
            <li class="last-episode">Last Episode Season: {{this.lastEpisodeToAir['season_number']}}</li>
          </div>
          <div class="show-div">
            <li class="homepage">Homepage: <a href={{this.homepage}}>{{this.homepage}}</a>
          </div>
          <div class="show-div">
            <li class="in-production" *ngIf="this.inProduction != 'false'">Is this Show in Production: Yes</li>
          </div>
          <div class="show-div">
            <li class="tagline">Tagline for {{show.name}}: {{this.tageline}}</li>
          </div>
          <div *ngIf="show!.video != undefined"a class="show-div">
            <li class="video">Is a video: {{show!.video}}</li>
          </div>
        </ul>

        <div class="download-quality">
          <select [(ngModel)]="quality" name="quality" id="quality">
            <option value="4k">4k</option>
            <option value="2k">2k</option>
            <option value="1080p">1080p</option>
            <option value="720p">720p</option>
            <option value="480p">480p</option>
            <option value="240p">240p</option>
          </select>
        </div>

        <button class="download-button" (click)="downloadShow(show.name, show.original_language, this.quality)">Download Show</button>
      </section>
    </article>
    `,
  styleUrls: ['./airing-today-tv-shows-details.component.sass'],
})

export class AiringTodayTvShowsDetailsComponent {
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
  public quality: string;
  public tmdbId: number = 0;

  constructor() {

  }

  ngOnInit() {
    const tvShowId = parseInt(this.route.snapshot.params['id'], 10);
    const page = parseInt(this.route.snapshot.params['page'], 10);
    const query = this.route.snapshot.params['query'];

    this.tvShowService.getAiringToday(page).subscribe((resp) => {
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
        let totalResult = resp['total_result'];

        let result: TvShowResult[] = [{adult: isAdult, backdrop_path: backdropPath, genre_ids: genreIds, id: id, name: name, first_air_date: firstAirDate, original_language: originalLanguage, original_name: originalName, overview: overview, popularity: popularity, poster_path: posterPath, vote_average: voteAverage, vote_count: voteCount, video: video}]

        this.tvShowList.push({ page: page, results: result,  total_pages: totalPages, total_result: totalResult });
      })

      this.tvShowList.splice(0, 1);

      for (var i = 0; i < this.tvShowList.length; i++) {;
        this.fetchedShow = this.tvShowList.find(movieResult => movieResult.results[i]!.id === tvShowId);
      }
    })

    this.tvShowService.getShowDetails(tvShowId).subscribe(show => {
      this.showsLength = show.seasons.length;
      for (var i = 0; i < this.showsLength; i++) {
        this.seasonEpisodeNumbers.push(show.seasons[i]['episode_count'])
        this.totalSeason.push(show.seasons[i]['season_number'])
        this.status = show.status
        this.createdBy.push(show.created_by[i])
        this.firstAirDate = show.first_air_date;
        this.homepage = show.homepage;
        this.inProduction = show.in_production;
        this.lastAirDate = show.last_air_date;
        this.lastEpisodeToAir = show.last_episode_to_air;
        this.tageline = show.tagline;
        this.tmdbId = show.id;
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
      console.log(this.episodes);
      this.tvShowService.makeAnimeDownloadRequest(title, this.episodes).subscribe(request => console.log(request))
    } else {
      console.log('TV');
      this.tvShowService.makeTvShowDownloadRequest(title, this.seasonEpisodeNumbers, this.quality, this.tmdbId).subscribe(request => console.log(request));
    }
  }
}
