import { ComponentFixture, TestBed } from '@angular/core/testing';

import { AddUrl } from './add-url';

describe('AddUrl', () => {
  let component: AddUrl;
  let fixture: ComponentFixture<AddUrl>;

  beforeEach(async () => {
    await TestBed.configureTestingModule({
      imports: [AddUrl]
    })
    .compileComponents();

    fixture = TestBed.createComponent(AddUrl);
    component = fixture.componentInstance;
    await fixture.whenStable();
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });
});
