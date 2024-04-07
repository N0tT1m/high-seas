import { ComponentFixture, TestBed } from '@angular/core/testing';

import { TopRatedTvShowsDetailsComponent } from './top-rated-tv-shows-details.component';

describe('TopRatedTvShowsDetailsComponent', () => {
  let component: TopRatedTvShowsDetailsComponent;
  let fixture: ComponentFixture<TopRatedTvShowsDetailsComponent>;

  beforeEach(async () => {
    await TestBed.configureTestingModule({
      imports: [TopRatedTvShowsDetailsComponent]
    })
    .compileComponents();
    
    fixture = TestBed.createComponent(TopRatedTvShowsDetailsComponent);
    component = fixture.componentInstance;
    fixture.detectChanges();
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });
});
