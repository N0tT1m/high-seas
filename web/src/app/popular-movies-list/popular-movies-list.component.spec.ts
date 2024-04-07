import { ComponentFixture, TestBed } from '@angular/core/testing';

import { PopularMoviesListComponent } from './popular-movies-list.component';

describe('PopularMoviesListComponent', () => {
  let component: PopularMoviesListComponent;
  let fixture: ComponentFixture<PopularMoviesListComponent>;

  beforeEach(async () => {
    await TestBed.configureTestingModule({
      imports: [PopularMoviesListComponent]
    })
    .compileComponents();
    
    fixture = TestBed.createComponent(PopularMoviesListComponent);
    component = fixture.componentInstance;
    fixture.detectChanges();
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });
});
