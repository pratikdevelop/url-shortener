import { ComponentFixture, TestBed } from '@angular/core/testing';

import { UrlLists } from './url-lists';

describe('UrlLists', () => {
  let component: UrlLists;
  let fixture: ComponentFixture<UrlLists>;

  beforeEach(async () => {
    await TestBed.configureTestingModule({
      imports: [UrlLists]
    })
    .compileComponents();

    fixture = TestBed.createComponent(UrlLists);
    component = fixture.componentInstance;
    await fixture.whenStable();
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });
});
