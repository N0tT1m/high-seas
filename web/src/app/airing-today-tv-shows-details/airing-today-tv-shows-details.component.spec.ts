import { ComponentFixture, TestBed } from '@angular/core/testing';

import { AiringTodayTvShowsDetailsComponent } from './airing-today-tv-shows-details.component';

describe('AiringTodayTvShowsDetailsComponent', () => {
  let component: AiringTodayTvShowsDetailsComponent;
  let fixture: ComponentFixture<AiringTodayTvShowsDetailsComponent>;

  beforeEach(async () => {
    await TestBed.configureTestingModule({
      imports: [AiringTodayTvShowsDetailsComponent]
    })
    .compileComponents();
    
    fixture = TestBed.createComponent(AiringTodayTvShowsDetailsComponent);
    component = fixture.componentInstance;
    fixture.detectChanges();
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });
});
