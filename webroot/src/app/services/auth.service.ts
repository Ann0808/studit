import { Injectable } from '@angular/core';
import { Http, Headers } from '@angular/http';
import { User } from './../components/pages/authorization/user'

@Injectable()
export class AuthService {
    isAuthenticated: boolean = false;
    constructor(private http: Http) { }

    authenticatenow(user: User) {
        var headers = new Headers();

        headers.append('Content-Type', 'application/x-www-form-urlencoded');

        return new Promise((resolve) => {
            this.http.post('http://localhost:8080/v1/auth/login/', JSON.stringify({Login: user.email, Password: user.password}), { headers: headers }).subscribe((data) => {
                if (data.json().success) {
                    window.localStorage.setItem('auth_key', data.json().token);
                    this.isAuthenticated = true;
                }
                resolve(this.isAuthenticated);
            }
            )

        })
    }
}