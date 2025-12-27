import { Routes } from '@angular/router';
import { Landing } from './landing/landing';
import SignupComponent from './auth/signup/signup';
import LoginComponent from './auth/login/login';
import { UrlLists } from './url-lists/url-lists';

export const routes: Routes = [
    {
        path: "",
        component: Landing
    },
    {
        path:"signup",
        component: SignupComponent
    },
    {
        path:"login",
        component: LoginComponent
    }, 
    {
        path:  "dashboard",
        component: UrlLists
    }
];
