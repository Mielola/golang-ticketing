CREATE TABLE user_logs (
    id INT AUTO_INCREMENT PRIMARY KEY,
    user_email VARCHAR(255) NULL, 
    login_time DATETIME NOT NULL,
    logout_time DATETIME NULL,
    shift_name VARCHAR(255) NULL,
    OTP VARCHAR(255) NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT fk_login_user FOREIGN KEY (user_email) REFERENCES users(email) ON DELETE CASCADE
);
