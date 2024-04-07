import { ComponentFixture, TestBed } from '@angular/core/testing';

import { NowPlayingMoviesDetailsComponent } from './now-playing-movies-details.component';

describe('NowPlayingMoviesDetailsComponent', () => {
  let component: NowPlayingMoviesDetailsComponent;
  let fixture: ComponentFixture<NowPlayingMoviesDetailsComponent>;

  beforeEach(async () => {
    await TestBed.configureTestingModule({
      imports: [NowPlayingMoviesDetailsComponent]
    })
    .compileComponents();
    
    fixture = TestBed.createComponent(NowPlayingMoviesDetailsComponent);
    component = fixture.componentInstance;
    fixture.detectChanges();
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });
});
