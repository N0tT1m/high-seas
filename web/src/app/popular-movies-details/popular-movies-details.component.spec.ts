import { ComponentFixture, TestBed } from '@angular/core/testing';

import { PopularMoviesDetailsComponent } from './popular-movies-details.component';

describe('PopularMoviesDetailsComponent', () => {
  let component: PopularMoviesDetailsComponent;
  let fixture: ComponentFixture<PopularMoviesDetailsComponent>;

  beforeEach(async () => {
    await TestBed.configureTestingModule({
      imports: [PopularMoviesDetailsComponent]
    })
    .compileComponents();
    
    fixture = TestBed.createComponent(PopularMoviesDetailsComponent);
    component = fixture.componentInstance;
    fixture.detectChanges();
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });
});
