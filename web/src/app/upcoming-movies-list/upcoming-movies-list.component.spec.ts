import { ComponentFixture, TestBed } from '@angular/core/testing';

import { UpcomingMoviesListComponent } from './upcoming-movies-list.component';

describe('UpcomingMoviesListComponent', () => {
  let component: UpcomingMoviesListComponent;
  let fixture: ComponentFixture<UpcomingMoviesListComponent>;

  beforeEach(async () => {
    await TestBed.configureTestingModule({
      imports: [UpcomingMoviesListComponent]
    })
    .compileComponents();
    
    fixture = TestBed.createComponent(UpcomingMoviesListComponent);
    component = fixture.componentInstance;
    fixture.detectChanges();
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });
});
