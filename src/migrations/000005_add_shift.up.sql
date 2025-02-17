CREATE TABLE shifts (
    id         INT PRIMARY KEY AUTO_INCREMENT,
    shift_name       VARCHAR(50) NOT NULL, 
    start_time TIME NOT NULL,        
    end_time   TIME NOT NULL,         
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);