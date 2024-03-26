import { TestBed } from '@angular/core/testing';

import { TvServiceService } from './tv-service.service';

describe('TvServiceService', () => {
  let service: TvServiceService;

  beforeEach(() => {
    TestBed.configureTestingModule({});
    service = TestBed.inject(TvServiceService);
  });

  it('should be created', () => {
    expect(service).toBeTruthy();
  });
});
