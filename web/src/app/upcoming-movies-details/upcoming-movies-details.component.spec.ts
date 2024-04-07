import { ComponentFixture, TestBed } from '@angular/core/testing';

import { UpcomingMoviesDetailsComponent } from './upcoming-movies-details.component';

describe('UpcomingMoviesDetailsComponent', () => {
  let component: UpcomingMoviesDetailsComponent;
  let fixture: ComponentFixture<UpcomingMoviesDetailsComponent>;

  beforeEach(async () => {
    await TestBed.configureTestingModule({
      imports: [UpcomingMoviesDetailsComponent]
    })
    .compileComponents();
    
    fixture = TestBed.createComponent(UpcomingMoviesDetailsComponent);
    component = fixture.componentInstance;
    fixture.detectChanges();
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });
});
