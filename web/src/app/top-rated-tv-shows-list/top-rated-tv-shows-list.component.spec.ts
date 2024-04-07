import { ComponentFixture, TestBed } from '@angular/core/testing';

import { TopRatedTvShowsListComponent } from './top-rated-tv-shows-list.component';

describe('TopRatedTvShowsListComponent', () => {
  let component: TopRatedTvShowsListComponent;
  let fixture: ComponentFixture<TopRatedTvShowsListComponent>;

  beforeEach(async () => {
    await TestBed.configureTestingModule({
      imports: [TopRatedTvShowsListComponent]
    })
    .compileComponents();
    
    fixture = TestBed.createComponent(TopRatedTvShowsListComponent);
    component = fixture.componentInstance;
    fixture.detectChanges();
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });
});
