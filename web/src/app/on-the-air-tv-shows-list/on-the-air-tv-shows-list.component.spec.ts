import { ComponentFixture, TestBed } from '@angular/core/testing';

import { OnTheAirTvShowsListComponent } from './on-the-air-tv-shows-list.component';

describe('OnTheAirTvShowsListComponent', () => {
  let component: OnTheAirTvShowsListComponent;
  let fixture: ComponentFixture<OnTheAirTvShowsListComponent>;

  beforeEach(async () => {
    await TestBed.configureTestingModule({
      imports: [OnTheAirTvShowsListComponent]
    })
    .compileComponents();
    
    fixture = TestBed.createComponent(OnTheAirTvShowsListComponent);
    component = fixture.componentInstance;
    fixture.detectChanges();
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });
});
