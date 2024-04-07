import { ComponentFixture, TestBed } from '@angular/core/testing';

import { TopRatedMoviesDetailsComponent } from './top-rated-movies-details.component';

describe('TopRatedMoviesDetailsComponent', () => {
  let component: TopRatedMoviesDetailsComponent;
  let fixture: ComponentFixture<TopRatedMoviesDetailsComponent>;

  beforeEach(async () => {
    await TestBed.configureTestingModule({
      imports: [TopRatedMoviesDetailsComponent]
    })
    .compileComponents();
    
    fixture = TestBed.createComponent(TopRatedMoviesDetailsComponent);
    component = fixture.componentInstance;
    fixture.detectChanges();
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });
});
