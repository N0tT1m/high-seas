import { ComponentFixture, TestBed } from '@angular/core/testing';

import { AiringTodayTvShowsListComponent } from './airing-today-tv-shows-list.component';

describe('AiringTodayTvShowsListComponent', () => {
  let component: AiringTodayTvShowsListComponent;
  let fixture: ComponentFixture<AiringTodayTvShowsListComponent>;

  beforeEach(async () => {
    await TestBed.configureTestingModule({
      imports: [AiringTodayTvShowsListComponent]
    })
    .compileComponents();
    
    fixture = TestBed.createComponent(AiringTodayTvShowsListComponent);
    component = fixture.componentInstance;
    fixture.detectChanges();
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });
});
