// src/app/features/settings/settings.component.ts
import { Component, OnInit } from '@angular/core';
import { CommonModule } from '@angular/common';
import { FormsModule } from '@angular/forms';

@Component({
  selector: 'app-settings',
  standalone: true,
  imports: [CommonModule, FormsModule],
  template: `
    <div class="settings-container">
      <h1>Settings</h1>

      <div class="settings-section">
        <h2>Plex Server</h2>
        <div class="setting-group">
          <label for="server-url">Server URL</label>
          <input
            type="text"
            id="server-url"
            [(ngModel)]="settings.serverUrl"
            placeholder="http://your-plex-server:32400"
          >
        </div>
      </div>

      <div class="settings-section">
        <h2>Sync Settings</h2>
        <div class="setting-group">
          <label class="toggle-label">
            <span>Enable Sync</span>
            <div class="toggle-switch">
              <input
                type="checkbox"
                [(ngModel)]="settings.syncEnabled"
              >
              <span class="toggle-slider"></span>
            </div>
          </label>
        </div>

        <div class="setting-group" *ngIf="settings.syncEnabled">
          <label for="sync-server">Sync Server URL</label>
          <input
            type="text"
            id="sync-server"
            [(ngModel)]="settings.syncServer"
            placeholder="http://your-sync-server:8080"
          >
        </div>

        <div class="setting-group" *ngIf="settings.syncEnabled">
          <label for="sync-interval">Sync Interval (seconds)</label>
          <input
            type="number"
            id="sync-interval"
            [(ngModel)]="settings.syncInterval"
            min="5"
            max="60"
          >
        </div>
      </div>

      <div class="settings-section">
        <h2>Appearance</h2>
        <div class="setting-group">
          <label for="theme">Theme</label>
          <select id="theme" [(ngModel)]="settings.theme">
            <option value="dark">Dark</option>
            <option value="darker">Darker</option>
            <option value="light">Light</option>
          </select>
        </div>

        <div class="setting-group">
          <label for="accent-color">Accent Color</label>
          <select id="accent-color" [(ngModel)]="settings.accentColor">
            <option value="orange">Orange</option>
            <option value="blue">Blue</option>
            <option value="green">Green</option>
            <option value="purple">Purple</option>
            <option value="red">Red</option>
          </select>
        </div>
      </div>

      <div class="settings-section">
        <h2>Playback</h2>
        <div class="setting-group">
          <label for="default-volume">Default Volume</label>
          <input
            type="range"
            id="default-volume"
            [(ngModel)]="settings.defaultVolume"
            min="0"
            max="100"
            step="1"
          >
          <span class="range-value">{{ settings.defaultVolume }}%</span>
        </div>

        <div class="setting-group">
          <label class="toggle-label">
            <span>Hardware Acceleration</span>
            <div class="toggle-switch">
              <input
                type="checkbox"
                [(ngModel)]="settings.hardwareAcceleration"
              >
              <span class="toggle-slider"></span>
            </div>
          </label>
        // Continuing the settings.component.ts template
        </div>

        <div class="setting-group">
          <label class="toggle-label">
            <span>Auto-Play Next Episode</span>
            <div class="toggle-switch">
              <input
                type="checkbox"
                [(ngModel)]="settings.autoPlayNext"
              >
              <span class="toggle-slider"></span>
            </div>
          </label>
        </div>
      </div>

      <div class="settings-section">
        <h2>About</h2>
        <div class="about-info">
          <p><strong>SyncFlex</strong> version 1.0.0</p>
          <p>A custom Plex client with synchronization capabilities</p>
        </div>
      </div>

      <div class="action-buttons">
        <button class="save-button" (click)="saveSettings()">Save Settings</button>
        <button class="reset-button" (click)="resetSettings()">Reset to Defaults</button>
      </div>
    </div>
  `,
  styles: [`
    .settings-container {
      padding: 1rem;
      max-width: 800px;
      margin: 0 auto;
    }

    h1 {
      font-size: 2rem;
      font-weight: 500;
      margin-bottom: 2rem;
      color: white;
    }

    .settings-section {
      background-color: #1f1f1f;
      border-radius: 8px;
      padding: 1.5rem;
      margin-bottom: 2rem;
    }

    h2 {
      font-size: 1.25rem;
      font-weight: 500;
      margin-top: 0;
      margin-bottom: 1.5rem;
      color: white;
      border-bottom: 1px solid #333;
      padding-bottom: 0.5rem;
    }

    .setting-group {
      margin-bottom: 1.5rem;
    }

    label {
      display: block;
      margin-bottom: 0.5rem;
      color: #bbb;
    }

    input[type="text"],
    input[type="number"],
    select {
      width: 100%;
      padding: 0.75rem;
      border: 1px solid #333;
      border-radius: 4px;
      background-color: #2a2a2a;
      color: white;
      font-size: 1rem;
    }

    input[type="range"] {
      width: calc(100% - 50px);
      margin-right: 10px;
    }

    .range-value {
      color: #bbb;
      width: 40px;
      display: inline-block;
      text-align: right;
    }

    .toggle-label {
      display: flex;
      justify-content: space-between;
      align-items: center;
    }

    .toggle-switch {
      position: relative;
      display: inline-block;
      width: 52px;
      height: 26px;
    }

    .toggle-switch input {
      opacity: 0;
      width: 0;
      height: 0;
    }

    .toggle-slider {
      position: absolute;
      cursor: pointer;
      top: 0;
      left: 0;
      right: 0;
      bottom: 0;
      background-color: #444;
      transition: .4s;
      border-radius: 26px;
    }

    .toggle-slider:before {
      position: absolute;
      content: "";
      height: 18px;
      width: 18px;
      left: 4px;
      bottom: 4px;
      background-color: white;
      transition: .4s;
      border-radius: 50%;
    }

    input:checked + .toggle-slider {
      background-color: #ff7b00;
    }

    input:checked + .toggle-slider:before {
      transform: translateX(26px);
    }

    .about-info {
      color: #bbb;
    }

    .action-buttons {
      display: flex;
      gap: 1rem;
      margin-top: 2rem;
    }

    .save-button, .reset-button {
      padding: 0.75rem 1.5rem;
      border: none;
      border-radius: 4px;
      font-size: 1rem;
      cursor: pointer;
      transition: background-color 0.2s ease;
    }

    .save-button {
      background-color: #ff7b00;
      color: white;
    }

    .save-button:hover {
      background-color: #e06e00;
    }

    .reset-button {
      background-color: #333;
      color: white;
    }

    .reset-button:hover {
      background-color: #444;
    }
  `]
})
export class SettingsComponent implements OnInit {
  settings = {
    serverUrl: '',
    syncEnabled: true,
    syncServer: '',
    syncInterval: 10,
    theme: 'dark',
    accentColor: 'orange',
    defaultVolume: 75,
    hardwareAcceleration: true,
    autoPlayNext: true
  };

  constructor() {}

  ngOnInit(): void {
    this.loadSettings();
  }

  loadSettings(): void {
    const savedSettings = localStorage.getItem('syncflex_settings');
    if (savedSettings) {
      this.settings = JSON.parse(savedSettings);
    }
  }

  saveSettings(): void {
    localStorage.setItem('syncflex_settings', JSON.stringify(this.settings));
    // In a real app, you might want to broadcast these changes to other components
    alert('Settings saved successfully!');
  }

  resetSettings(): void {
    this.settings = {
      serverUrl: '',
      syncEnabled: true,
      syncServer: '',
      syncInterval: 10,
      theme: 'dark',
      accentColor: 'orange',
      defaultVolume: 75,
      hardwareAcceleration: true,
      autoPlayNext: true
    };
  }
}
