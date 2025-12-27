// components/url-shortener/url-shortener.component.ts
import { Component, OnInit } from '@angular/core';
import { FormBuilder, FormGroup, FormsModule, ReactiveFormsModule, Validators } from '@angular/forms';
import { MatSnackBar, MatSnackBarModule } from '@angular/material/snack-bar';
import { Clipboard, ClipboardModule } from '@angular/cdk/clipboard';
import { HttpClientModule } from '@angular/common/http';
import { MatButtonModule } from '@angular/material/button';
import { MatCardModule } from '@angular/material/card';
import { MatIconModule } from '@angular/material/icon';
import { MatInputModule } from '@angular/material/input';
import { MatProgressSpinnerModule } from '@angular/material/progress-spinner';
import { MatToolbarModule } from '@angular/material/toolbar';
import { MatDividerModule } from '@angular/material/divider';
import { MatListModule } from '@angular/material/list';
import { MatTabsModule } from '@angular/material/tabs';
import { CommonModule } from '@angular/common';
import { RouterModule } from '@angular/router';
interface ShortenedUrl {
  original: string;
  shortened: string;
  createdAt: Date;
  clicks: number;
}

@Component({
  selector: 'app-landing',
  imports: [ CommonModule,
    HttpClientModule,
    FormsModule,
    ReactiveFormsModule,
    ClipboardModule,
    
    // Angular Material
    MatButtonModule,
    MatInputModule,
    MatIconModule,
    MatToolbarModule,
    MatCardModule,
    MatSnackBarModule,
    MatProgressSpinnerModule,
    MatDividerModule,
    MatListModule,
    MatTabsModule,
    RouterModule
  ],
  templateUrl: './landing.html',
  styleUrl: './landing.css',
})
export class Landing {


  urlForm: FormGroup;
  isGenerating = false;
  shortenedUrl = '';
  recentUrls: ShortenedUrl[] = [];
  activeTab = 0;

  // Stats
  totalUrls = 1234;
  totalClicks = 56789;
  todayClicks = 234;

  constructor(
    private fb: FormBuilder,
    private snackBar: MatSnackBar,
    private clipboard: Clipboard
  ) {
    this.urlForm = this.fb.group({
      url: ['', [Validators.required, Validators.pattern('https?://.+')]],
      customAlias: ['', [Validators.pattern('^[a-zA-Z0-9-_]+$')]]
    });
  }

  ngOnInit(): void {
    // Load recent URLs from localStorage or API
    this.loadRecentUrls();
  }

  loadRecentUrls(): void {
    // Mock data - replace with actual API call
    this.recentUrls = [
      { original: 'https://example.com/very-long-url-path-that-needs-shortening', shortened: 'https://short.ly/abc123', createdAt: new Date(), clicks: 45 },
      { original: 'https://github.com/angular/angular', shortened: 'https://short.ly/ng-awesome', createdAt: new Date(Date.now() - 86400000), clicks: 120 },
      { original: 'https://stackoverflow.com/questions/tagged/angular', shortened: 'https://short.ly/so-help', createdAt: new Date(Date.now() - 172800000), clicks: 89 },
    ];
  }

  generateShortUrl(): void {
    if (this.urlForm.invalid) {
      this.snackBar.open('Please enter a valid URL starting with http:// or https://', 'OK', {
        duration: 3000,
        panelClass: ['error-snackbar']
      });
      return;
    }

    this.isGenerating = true;
    const url = this.urlForm.get('url')?.value;
    const customAlias = this.urlForm.get('customAlias')?.value;

    // Simulate API call
    setTimeout(() => {
      const randomId = customAlias || this.generateRandomId(6);
      this.shortenedUrl = `https://short.ly/${randomId}`;
      
      // Add to recent URLs
      const newUrl: ShortenedUrl = {
        original: url,
        shortened: this.shortenedUrl,
        createdAt: new Date(),
        clicks: 0
      };
      
      this.recentUrls.unshift(newUrl);
      if (this.recentUrls.length > 5) {
        this.recentUrls.pop();
      }
      
      this.isGenerating = false;
      this.activeTab = 1; // Switch to results tab
      
      this.snackBar.open('URL shortened successfully!', 'OK', {
        duration: 2000,
        panelClass: ['success-snackbar']
      });
    }, 1000);
  }

  copyToClipboard(text: string): void {
    this.clipboard.copy(text);
    this.snackBar.open('Copied to clipboard!', 'OK', {
      duration: 2000,
      panelClass: ['success-snackbar']
    });
  }

  generateRandomId(length: number): string {
    const chars = 'abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789';
    let result = '';
    for (let i = 0; i < length; i++) {
      result += chars.charAt(Math.floor(Math.random() * chars.length));
    }
    return result;
  }

  formatDate(date: Date): string {
    return date.toLocaleDateString('en-US', {
      month: 'short',
      day: 'numeric',
      year: 'numeric'
    });
  }

  resetForm(): void {
    this.urlForm.reset();
    this.shortenedUrl = '';
    this.activeTab = 0;
  }
}