import { ComponentFixture, TestBed } from '@angular/core/testing';

import { PopularTvShowsDetailsComponent } from './popular-tv-shows-details.component';

describe('PopularTvShowsDetailsComponent', () => {
  let component: PopularTvShowsDetailsComponent;
  let fixture: ComponentFixture<PopularTvShowsDetailsComponent>;

  beforeEach(async () => {
    await TestBed.configureTestingModule({
      imports: [PopularTvShowsDetailsComponent]
    })
    .compileComponents();
    
    fixture = TestBed.createComponent(PopularTvShowsDetailsComponent);
    component = fixture.componentInstance;
    fixture.detectChanges();
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });
});
