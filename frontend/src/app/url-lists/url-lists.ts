import { ClipboardModule } from '@angular/cdk/clipboard';
import { CommonModule } from '@angular/common';
import { HttpClient } from '@angular/common/http';
import { Component, inject, signal } from '@angular/core';
import { Field } from '@angular/forms/signals';
import { MatButtonModule } from '@angular/material/button';
import { MatCardModule } from '@angular/material/card';
import { MatNativeDateModule } from '@angular/material/core';
import { MatDatepickerModule } from '@angular/material/datepicker';
import { MatDialogModule, MatDialog } from '@angular/material/dialog';
import { MatFormFieldModule } from '@angular/material/form-field';
import { MatIconModule } from '@angular/material/icon';
import { MatInputModule } from '@angular/material/input';
import { MatProgressSpinnerModule } from '@angular/material/progress-spinner';
import { MatSnackBarModule, MatSnackBar } from '@angular/material/snack-bar';
import { MatTableModule } from '@angular/material/table';
import { MatToolbarModule } from '@angular/material/toolbar';
import { MatTooltipModule } from '@angular/material/tooltip';
import { AddUrl } from './add-url/add-url';
import { Router } from '@angular/router';

interface ShortenedURL {
  original_url: string;
  short_code: string;
  title?: string;
  created_at?: string;
  expires_at?: string | null;
  click_count?: number;
}

@Component({
  selector: 'app-url-lists',
  standalone: true,
  imports: [
    CommonModule,
    MatCardModule,
    MatTableModule,
    MatToolbarModule,
    MatIconModule,
    MatButtonModule,
    MatTooltipModule,
    MatProgressSpinnerModule,
    ClipboardModule,
    MatDialogModule,
    MatFormFieldModule,
    MatInputModule,
    MatDatepickerModule,
    MatNativeDateModule,
    MatSnackBarModule,
  ],
  templateUrl: './url-lists.html',
  styleUrl: './url-lists.css',
})
export class UrlLists {
  private readonly http = inject(HttpClient);
  private readonly dialog = inject(MatDialog);
  private readonly snackBar = inject(MatSnackBar);
  private readonly router =  inject(Router)

  title = signal('URL Shortener');
  urls = signal<ShortenedURL[]>([]);
  loading = signal(true);
  submitting = signal(false);
  error = signal<string | null>(null);

  // For custom alias availability
  customAliasTaken = signal(false);
  checkingAlias = signal(false);

  displayedColumns: string[] = [
    'index',
    'title',
    'originalUrl',
    'shortCode',
    'expiresAt',
    'actions',
    'clicks'
  ];

  ngOnInit(): void {
    this.loadUrls();
  }

  loadUrls(): void {
    this.loading.set(true);
    this.error.set(null);

    this.http.get<ShortenedURL[]>('http://localhost:8080/api/urls').subscribe({
      next: (data) => {
        this.urls.set(data);
        this.loading.set(false);
      },
      error: () => {
        this.error.set('Failed to load URLs');
        this.loading.set(false);
      },
    });
  }

openAddUrlDialog(): void {
  const dialogRef = this.dialog.open(AddUrl, {
    width: '500px',
    autoFocus: false,
  });

  dialogRef.afterClosed().subscribe(result => {
    if (result === 'success') {  // We'll make AddUrl return this
      this.loadUrls(); // Refresh the list
    }
  });
}
  getShortUrl(shortCode: string): string {
    return `http://localhost:8080/${shortCode}`;
  }

  onCopied(success: boolean): void {
    if (success) {
      this.snackBar.open('Copied to clipboard!', 'Close', { duration: 2000 });
    }
  }
  deleteUrl(shortCode: string) {
  if (!confirm('Are you sure you want to delete this URL?')) return;

  this.http.delete(`http://localhost:8080/api/url/${shortCode}`).subscribe({
    next: () => {
      this.urls.update(list => list.filter(u => u.short_code !== shortCode));
      this.snackBar.open('URL deleted', 'Close', { duration: 3000 });
    },
    error: () => this.snackBar.open('Failed to delete', 'Close', { duration: 3000 })
  });
}
openEditDialog(url: ShortenedURL): void {
  const dialogRef = this.dialog.open(AddUrl, {
    width: '500px',
    autoFocus: false,
    data: {  // Pass the existing URL data to the dialog
      mode: 'edit',
      url: url
    }
  });

  dialogRef.afterClosed().subscribe(result => {
    if (result === 'success') {
      this.loadUrls(); // Refresh the list after edit
    }
  });
}
logout(): void {
  // Remove the JWT token (adjust key if you use a different one)
  localStorage.removeItem('token');

  // Optional: show feedback
  this.snackBar.open('Logged out successfully', 'Close', { duration: 3000 });

  // Redirect to login page or reload (depending on your routing)
  // If you have a separate login route:
  this.router.navigate(['/']);
}
}
