// tv-service.service.spec.ts
import { TestBed } from '@angular/core/testing';
import { TvShowService } from './tv-service.service';

describe('TvService', () => {
  let service: TvShowService;

  beforeEach(() => {
    TestBed.configureTestingModule({});
    service = TestBed.inject(TvShowService);
  });

  it('should be created', () => {
    expect(service).toBeTruthy();
  });
});
