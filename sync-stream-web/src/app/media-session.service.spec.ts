import { TestBed } from '@angular/core/testing';

import { MediasessionService } from './media-session.service';

describe('MediasessionService', () => {
  let service: MediasessionService;

  beforeEach(() => {
    TestBed.configureTestingModule({});
    service = TestBed.inject(MediasessionService);
  });

  it('should be created', () => {
    expect(service).toBeTruthy();
  });
});
