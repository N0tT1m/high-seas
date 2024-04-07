import { ComponentFixture, TestBed } from '@angular/core/testing';

import { UpcomingMoviesComponent } from './upcoming-movies.component';

describe('UpcomingMoviesComponent', () => {
  let component: UpcomingMoviesComponent;
  let fixture: ComponentFixture<UpcomingMoviesComponent>;

  beforeEach(async () => {
    await TestBed.configureTestingModule({
      imports: [UpcomingMoviesComponent]
    })
    .compileComponents();
    
    fixture = TestBed.createComponent(UpcomingMoviesComponent);
    component = fixture.componentInstance;
    fixture.detectChanges();
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });
});
