import { HttpClient, httpResource } from '@angular/common/http';
import { Component, inject, OnInit, signal } from '@angular/core';
import { MatCardModule } from '@angular/material/card';
import { MatListModule } from '@angular/material/list';
import { MatIconModule } from '@angular/material/icon';
import { MatButtonModule } from '@angular/material/button';
import { MatTooltipModule } from '@angular/material/tooltip';
import { MatProgressSpinnerModule } from '@angular/material/progress-spinner';
import { ClipboardModule } from '@angular/cdk/clipboard';
import { CommonModule } from '@angular/common';
import { MatTableModule } from '@angular/material/table';

import { ViewChild, TemplateRef } from '@angular/core';
import { MatDialog, MatDialogModule, MatDialogRef } from '@angular/material/dialog';
import { MatToolbarModule } from '@angular/material/toolbar';
@Component({
  selector: 'app-root',
  imports: [
    MatIconModule,
    MatCardModule,
    MatButtonModule,
    MatTooltipModule,
    MatProgressSpinnerModule,
    ClipboardModule,
    CommonModule,
    MatTableModule,
    MatToolbarModule,
    MatDialogModule
  ],
  templateUrl: './app.html',
  styleUrl: './app.css',
})
export class App implements OnInit {
  protected readonly title = signal('frontend');
  private readonly http = inject(HttpClient);
  urls = signal<any>([]);
  // In your App component
  loading = signal(true);
  error = signal<string | null>(null);
  displayedColumns: string[] = ['index', 'name', 'originalUrl', 'shortCode', 'actions'];
  @ViewChild('myDialogTemplate') dialogTemplate!: TemplateRef<any>;
  private dialogRef!: MatDialogRef<any>;
  private readonly dialog = inject(MatDialog);
  onCopied(copied: boolean) {
    if (copied) {
      // Optional: show a toast/snackbar
      console.log('URL copied!');
    }
  }

  ngOnInit(): void {
    this.http.get('http://localhost:8080/urls').subscribe({
      next: (val: any) => {
        this.urls.set(val);
        this.loading.set(false);
      },
      error: (err) => {
        this.loading.set(false);

        console.error('h;ktl;ktf', err);
      },
      complete: () => {
        this.loading.set(false);
      },
    });
  }
  openAddUrlDialog(): void {
    console.log(
      "gfkgkjd"
    );
    this.dialogRef = this.dialog.open(this.dialogTemplate);
    
  }

  closeDialog (): void {
    this.dialogRef.close();
  }
}
