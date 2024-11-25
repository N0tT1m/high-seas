import { Component, Input, Output, EventEmitter } from '@angular/core';

@Component({
  selector: 'app-dropdown',
  templateUrl: './app-dropdown.component.html',
  standalone: true,
  styleUrls: ['./app-dropdown.component.css']
})
export class DropdownComponent {
  @Input() title: string;
  @Input() links: any[];
  @Input() dropdownOpen: boolean;
  @Output() dropdownOpenChange = new EventEmitter<boolean>();

  toggleDropdown(event: Event) {
    event.stopPropagation();
    this.dropdownOpen = !this.dropdownOpen;
    this.dropdownOpenChange.emit(this.dropdownOpen);
  }
}
