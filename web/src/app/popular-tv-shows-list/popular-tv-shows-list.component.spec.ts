import { ComponentFixture, TestBed } from '@angular/core/testing';

import { PopularTvShowsListComponent } from './popular-tv-shows-list.component';

describe('PopularTvShowsListComponent', () => {
  let component: PopularTvShowsListComponent;
  let fixture: ComponentFixture<PopularTvShowsListComponent>;

  beforeEach(async () => {
    await TestBed.configureTestingModule({
      imports: [PopularTvShowsListComponent]
    })
    .compileComponents();
    
    fixture = TestBed.createComponent(PopularTvShowsListComponent);
    component = fixture.componentInstance;
    fixture.detectChanges();
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });
});
