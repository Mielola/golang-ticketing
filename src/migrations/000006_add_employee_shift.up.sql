CREATE TABLE employee_shifts (
    id          INT PRIMARY KEY AUTO_INCREMENT,
    user_email  VARCHAR(255) NOT NULL,
    shift_id    BIGINT NOT NULL,
    shift_date  DATE NOT NULL,  
    created_at  TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_email) REFERENCES users(email) ON DELETE CASCADE,
    FOREIGN KEY (shift_id) REFERENCES shifts(id) ON DELETE CASCADE
);