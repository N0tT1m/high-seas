import { ComponentFixture, TestBed } from '@angular/core/testing';

import { AiringTodayTvShowsComponent } from './airing-today-tv-shows.component';

describe('AiringTodayTvShowsComponent', () => {
  let component: AiringTodayTvShowsComponent;
  let fixture: ComponentFixture<AiringTodayTvShowsComponent>;

  beforeEach(async () => {
    await TestBed.configureTestingModule({
      imports: [AiringTodayTvShowsComponent]
    })
    .compileComponents();
    
    fixture = TestBed.createComponent(AiringTodayTvShowsComponent);
    component = fixture.componentInstance;
    fixture.detectChanges();
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });
});
