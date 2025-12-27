import { CommonModule } from '@angular/common';
import { Component, OnInit } from '@angular/core';

// Signals-based form imports
import { RouterModule } from '@angular/router';


@Component({
  selector: 'app-root',
  standalone: true,
  imports: [
    CommonModule,
    RouterModule
  ],
  templateUrl: './app.html',
  styleUrl: './app.css',
})
export class App {
  
}
