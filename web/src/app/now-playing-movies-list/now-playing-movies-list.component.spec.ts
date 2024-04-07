import { ComponentFixture, TestBed } from '@angular/core/testing';

import { NowPlayingMoviesListComponent } from './now-playing-movies-list.component';

describe('NowPlayingMoviesListComponent', () => {
  let component: NowPlayingMoviesListComponent;
  let fixture: ComponentFixture<NowPlayingMoviesListComponent>;

  beforeEach(async () => {
    await TestBed.configureTestingModule({
      imports: [NowPlayingMoviesListComponent]
    })
    .compileComponents();
    
    fixture = TestBed.createComponent(NowPlayingMoviesListComponent);
    component = fixture.componentInstance;
    fixture.detectChanges();
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });
});
