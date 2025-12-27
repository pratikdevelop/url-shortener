import { Component, inject, Inject, OnInit, Optional, signal } from '@angular/core';
import { MAT_DIALOG_DATA, MatDialogModule, MatDialogRef } from '@angular/material/dialog';
import { HttpClient } from '@angular/common/http';
import { MatSnackBar, MatSnackBarModule } from '@angular/material/snack-bar';
import { ClipboardModule } from '@angular/cdk/clipboard';
import { CommonModule } from '@angular/common';
import { Field, form, maxLength, minLength, pattern, required } from '@angular/forms/signals';
import { MatButtonModule } from '@angular/material/button';
import { MatCardModule } from '@angular/material/card';
import { MatNativeDateModule } from '@angular/material/core';
import { MatDatepickerModule } from '@angular/material/datepicker';
import { MatFormFieldModule } from '@angular/material/form-field';
import { MatIconModule } from '@angular/material/icon';
import { MatInputModule } from '@angular/material/input';
import { MatProgressSpinnerModule } from '@angular/material/progress-spinner';
import { MatTableModule } from '@angular/material/table';
import { MatToolbarModule } from '@angular/material/toolbar';
import { MatTooltipModule } from '@angular/material/tooltip';

interface DialogData {
  mode: 'add' | 'edit';
  url?: ShortenedURL;
}

interface ShortenedURL {
  original_url: string;
  short_code: string;
  title?: string;
  expires_at?: unknown;
}

@Component({
  selector: 'app-add-url',
  standalone: true,
  imports: [
    MatDialogModule,
    MatButtonModule,
    MatFormFieldModule,
    MatInputModule,
    MatDatepickerModule,
    MatNativeDateModule,
    Field,
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
  templateUrl: './add-url.html',
})
export class AddUrl  implements  OnInit{
  private readonly dialogRef = inject(MatDialogRef<AddUrl>);
  private readonly http = inject(HttpClient);
  private readonly snackBar = inject(MatSnackBar);
  private readonly data: any = inject<DialogData>(MAT_DIALOG_DATA)


  mode: 'add' | 'edit' = 'add';
  shortCode = '';
  submitting = false;
  loading = signal(true);
  error = signal<string | null>(null);

  // For custom alias availability
  customAliasTaken = signal(false);
  checkingAlias = signal(false);


  // Form fields (you can use signals or simple variables)
   urlFormModel = signal({
    title: '',
    customAlias: '',
    originalUrl: '',
    expiryDate:  null as Date | null
  });

  urlForm = form(this.urlFormModel, (path) => {
    required(path.originalUrl, { message: 'Original URL is required' });
    pattern(path.originalUrl, /^https?:\/\/.+/, { message: 'Must start with http:// or https://' });

    if (path.customAlias) {
      minLength(path.customAlias, 3, { message: 'At least 3 characters' });
      maxLength(path.customAlias, 30, { message: 'Max 30 characters' });
      pattern(path.customAlias, /^[a-zA-Z0-9_-]+$/, { message: 'Only letters, numbers, _, - allowed' });
    }
  });
  title = '';
  originalUrl = '';
  customAlias = '';
  expiryDate:  Date | null = null ;

  ngOnInit(): void {
    console.log(
      this.data
    );
    
     if (this.data?.mode === 'edit' && this.data.url) {
      this.mode = 'edit';
      this.shortCode = this.data.url.short_code;
      this.title = this.data.url.title || '';
      this.originalUrl = this.data.url.original_url;
      this.customAlias = this.data.url.short_code; // Show but disable editing alias
      if (this.data.url.expires_at) {
        this.expiryDate = new Date(this.data.url.expires_at);
      }
      this.urlForm().setControlValue({
        title: this.title,
        originalUrl: this.data.url.original_url,
        customAlias: this.data.url.short_code,
        expiryDate: this.data.url.expires_at
      
        
      })   }
  }
  submit(): void {
    if (this.originalUrl.trim() === '') {
      this.snackBar.open('Original URL is required', 'Close', { duration: 4000 });
      return;
    }

    this.submitting = true;

    const payload: any = {
      original_url: this.originalUrl.trim(),
      title: this.title.trim() || undefined,
    };

    if (this.expiryDate) {
      payload.expires_at = this.expiryDate.toISOString();
    } else if (this.mode === 'edit') {
      payload.expires_at = null; // Clear expiry
    }

    const url = this.mode === 'add'
      ? 'http://localhost:8080/api/add-url'
      : `http://localhost:8080/api/url/${this.shortCode}`;

    const method = this.mode === 'add' ? 'post' : 'put';

    this.http.request(method, url, { body: payload }).subscribe({
      next: () => {
        this.snackBar.open(
          this.mode === 'add' ? 'URL shortened!' : 'URL updated!',
          'Close',
          { duration: 4000 }
        );
        this.dialogRef.close('success');
      },
      error: (err) => {
        const msg = err.error?.message || 'Operation failed';
        this.snackBar.open(msg, 'Close', { duration: 6000 });
        this.submitting = false;
      }
    });
  }

  close(): void {
    this.dialogRef.close();
  }
}