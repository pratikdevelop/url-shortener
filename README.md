```markdown
# URL Shortener Application

This is a full-stack URL shortener application with a **Go backend** and an **Angular frontend**.

## Features

- Shorten long URLs into concise, manageable links.
- Option to create **custom aliases** (short codes) for personalized URLs.
- User authentication (register & login with JWT).
- View a personalized list of all your shortened URLs.
- **Edit** title, original URL, or expiry date of existing links.
- **Delete** shortened URLs.
- **Click tracking** – see how many times each link was clicked.
- Public statistics page for any short link (`/s/{code}`).
- Protected analytics API for owners.
- Rate limiting and CORS protection.

## Technologies Used

### Backend (Go)

- **Go** – Core programming language.
- **net/http** – Standard library HTTP server (no external frameworks).
- **MongoDB** (via official driver) – NoSQL database for storing users and URLs.
- **JWT** (golang-jwt) – Authentication tokens.
- **bcrypt** – Password hashing.
- **tollbooth** – Rate limiting for DDoS/abuse protection.
- **rs/cors** – CORS middleware.
- **godotenv** – Environment variable loading.

### Frontend (Angular)

- **Angular** (latest standalone components) – Modern TypeScript framework.
- **Angular Material** – UI components (table, dialog, toolbar, etc.).
- **RxJS & Signals** – Reactive state management.
- **Clipboard CDK** – Copy-to-clipboard functionality.
- **Tailwind CSS** (optional) or Angular Material styling.

## Project Structure

```
url-shortener/
├── main.go                  # Complete Go backend (single file)
├── angular-app/             # Angular frontend project
│   ├── src/
│   │   ├── app/
│   │   │   ├── url-lists.component.ts/html/css
│   │   │   └── add-url-dialog.component.ts/html/css
│   │   └── ...
│   └── angular.json, package.json, etc.
```

## Setup and Installation

### 1. Clone the Repository

```bash
git clone https://github.com/pratikdevelop/url-shortener
cd url-shortener
```

### 2. Backend Setup (Go)

The backend is a single `main.go` file.

```bash
# From project root
go mod tidy
```

Create a `.env` file in the root directory:

```
MONGODB_URI=mongodb+srv://<user>:<password>@cluster0.mongodb.net/url-shortner?retryWrites=true&w=majority
```

> Replace with your actual MongoDB Atlas connection string.

### 3. Frontend Setup (Angular)

```bash
cd angular-app
npm install
```

## Running the Application

### 1. Run the Backend

```bash
go run main.go
```

Server starts at **http://localhost:8080**

### 2. Run the Frontend

```bash
cd angular-app
ng serve
```

Frontend available at **http://localhost:4200**

## API Endpoints (Base URL: http://localhost:8080)

### Authentication
- `POST /api/register` – Register new user
- `POST /api/login` – Login and receive JWT token

### URL Management (Protected – requires Bearer token)
- `POST /api/add-url` – Create new short URL  
  Body: `{ "original_url": "...", "custom_alias": "optional", "title": "optional", "expires_at": "ISO string optional" }`
- `GET /api/urls` – List all your URLs (with click_count)
- `PUT /api/url/{shortCode}` – Update title, original_url, or expires_at
- `DELETE /api/url/{shortCode}` – Delete a URL
- `GET /api/stats/{shortCode}` – Get detailed stats for one of your links

### Public Endpoints
- `GET /{shortCode}` – Redirect to original URL (increments click count)
- `GET /s/{shortCode}` – Public HTML statistics page (clicks, expiry, etc.)

## Security & Protection

- JWT-based authentication
- Rate limiting on all endpoints
- Unique constraints on email & short_code
- Input validation
- CORS restricted to frontend origin

## Future Improvements (Optional)

- QR code generation for short links
- Password reset flow
- User profile page
- Admin dashboard
- Deployment with Docker + Nginx + HTTPS

---

**Enjoy your own private, secure, and feature-rich URL shortener!**  
Built with modern Go and Angular – fast, clean, and ready for production.
```