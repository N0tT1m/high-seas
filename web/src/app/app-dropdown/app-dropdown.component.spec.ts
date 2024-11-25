import { ComponentFixture, TestBed } from '@angular/core/testing';

import { AppDropdownComponent } from './app-dropdown.component';

describe('AppDropdownComponent', () => {
  let component: AppDropdownComponent;
  let fixture: ComponentFixture<AppDropdownComponent>;

  beforeEach(async () => {
    await TestBed.configureTestingModule({
      imports: [AppDropdownComponent]
    })
    .compileComponents();
    
    fixture = TestBed.createComponent(AppDropdownComponent);
    component = fixture.componentInstance;
    fixture.detectChanges();
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });
});
