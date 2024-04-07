import { ComponentFixture, TestBed } from '@angular/core/testing';

import { OnTheAirTvShowsDetailsComponent } from './on-the-air-tv-shows-details.component';

describe('OnTheAirTvShowsDetailsComponent', () => {
  let component: OnTheAirTvShowsDetailsComponent;
  let fixture: ComponentFixture<OnTheAirTvShowsDetailsComponent>;

  beforeEach(async () => {
    await TestBed.configureTestingModule({
      imports: [OnTheAirTvShowsDetailsComponent]
    })
    .compileComponents();
    
    fixture = TestBed.createComponent(OnTheAirTvShowsDetailsComponent);
    component = fixture.componentInstance;
    fixture.detectChanges();
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });
});
