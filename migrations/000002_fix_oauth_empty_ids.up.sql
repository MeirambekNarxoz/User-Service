-- Replace empty OAuth IDs with NULL so unique indexes do not block email/password signups.
UPDATE users SET google_id = NULL WHERE google_id = '';
UPDATE users SET github_id = NULL WHERE github_id = '';
UPDATE users SET linkedin_id = NULL WHERE linkedin_id = '';
