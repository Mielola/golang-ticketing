-- phpMyAdmin SQL Dump
-- version 5.2.1
-- https://www.phpmyadmin.net/
--
-- Host: localhost:3306
-- Generation Time: Feb 28, 2025 at 02:38 AM
-- Server version: 8.4.3
-- PHP Version: 8.3.16

SET SQL_MODE = "NO_AUTO_VALUE_ON_ZERO";
START TRANSACTION;
SET time_zone = "+00:00";


/*!40101 SET @OLD_CHARACTER_SET_CLIENT=@@CHARACTER_SET_CLIENT */;
/*!40101 SET @OLD_CHARACTER_SET_RESULTS=@@CHARACTER_SET_RESULTS */;
/*!40101 SET @OLD_COLLATION_CONNECTION=@@COLLATION_CONNECTION */;
/*!40101 SET NAMES utf8mb4 */;

--
-- Database: `commandcenter`
--

-- --------------------------------------------------------

--
-- Table structure for table `category`
--

CREATE TABLE `category` (
  `id` bigint NOT NULL,
  `category_name` varchar(255) COLLATE utf8mb4_general_ci NOT NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;

--
-- Dumping data for table `category`
--

INSERT INTO `category` (`id`, `category_name`) VALUES
(1, 'Kendala Login'),
(2, 'Kesalahan Data'),
(3, 'Kesalahan Dokumen');

-- --------------------------------------------------------

--
-- Table structure for table `employee_shifts`
--

CREATE TABLE `employee_shifts` (
  `id` bigint NOT NULL,
  `user_email` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL,
  `shift_id` bigint NOT NULL,
  `shift_date` date NOT NULL,
  `created_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;

--
-- Dumping data for table `employee_shifts`
--

INSERT INTO `employee_shifts` (`id`, `user_email`, `shift_id`, `shift_date`, `created_at`) VALUES
(25, 'john.doe@example.com', 3, '2025-02-01', '2025-01-31 04:04:36'),
(26, 'john.doe@example.com', 2, '2025-01-31', '2025-01-31 04:39:17'),
(27, 'mwildab15@gmail.com', 2, '2025-02-19', '2025-01-31 07:21:36');

-- --------------------------------------------------------

--
-- Table structure for table `note`
--

CREATE TABLE `note` (
  `id` bigint NOT NULL,
  `Title` varchar(256) COLLATE utf8mb4_general_ci NOT NULL,
  `Content` varchar(256) COLLATE utf8mb4_general_ci NOT NULL,
  `user_email` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci DEFAULT NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;

--
-- Dumping data for table `note`
--

INSERT INTO `note` (`id`, `Title`, `Content`, `user_email`) VALUES
(25, 'Note Title', 'This is the content of the note.', 'john.doe@example.com');

-- --------------------------------------------------------

--
-- Table structure for table `products`
--

CREATE TABLE `products` (
  `id` bigint NOT NULL,
  `name` varchar(255) COLLATE utf8mb4_general_ci NOT NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;

--
-- Dumping data for table `products`
--

INSERT INTO `products` (`id`, `name`) VALUES
(1, 'Gugus Pangan'),
(2, 'Photobooth');

-- --------------------------------------------------------

--
-- Table structure for table `shifts`
--

CREATE TABLE `shifts` (
  `id` bigint NOT NULL,
  `shift_name` varchar(50) COLLATE utf8mb4_general_ci DEFAULT NULL,
  `start_time` time DEFAULT NULL,
  `end_time` time DEFAULT NULL,
  `created_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;

--
-- Dumping data for table `shifts`
--

INSERT INTO `shifts` (`id`, `shift_name`, `start_time`, `end_time`, `created_at`) VALUES
(1, 'Pagi', '08:00:00', '15:00:00', '2025-01-30 03:17:16'),
(2, 'Sore', '15:00:00', '23:00:00', '2025-01-30 03:28:05'),
(3, 'Malam', '23:00:00', '07:00:00', '2025-01-30 03:28:47');

-- --------------------------------------------------------

--
-- Table structure for table `shift_logs`
--

CREATE TABLE `shift_logs` (
  `id` bigint NOT NULL,
  `user_email` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `shift_id` bigint NOT NULL,
  `shift_date` date NOT NULL,
  `reason` text COLLATE utf8mb4_general_ci NOT NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;

--
-- Dumping data for table `shift_logs`
--

INSERT INTO `shift_logs` (`id`, `user_email`, `shift_id`, `shift_date`, `reason`) VALUES
(25, 'john.doe@example.com', 2, '2025-02-01', 'Sakit'),
(26, 'john.doe@example.com', 1, '2025-02-01', 'Sakit'),
(27, 'john.doe@example.com', 1, '2025-02-01', 'Sakit nihhh'),
(28, 'john.doe@example.com', 3, '2025-02-01', 'Sakit nihhh'),
(29, 'wildan1@example.com', 3, '2025-02-01', 'Sakit nihhh'),
(30, 'wildan1@example.com', 2, '2025-02-01', 'Tidak bisa hadir dikarenakan suatu keperluan');

-- --------------------------------------------------------

--
-- Table structure for table `tenants`
--

CREATE TABLE `tenants` (
  `id` bigint UNSIGNED NOT NULL,
  `name` varchar(191) COLLATE utf8mb4_general_ci DEFAULT NULL,
  `database` longtext COLLATE utf8mb4_general_ci
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;

--
-- Dumping data for table `tenants`
--

INSERT INTO `tenants` (`id`, `name`, `database`) VALUES
(1, 'Tenant1', 'tenant_Tenant1'),
(2, 'Whoops', 'tenant_Whoops'),
(3, 'Whoops1', 'tenants_Whoops1');

-- --------------------------------------------------------

--
-- Table structure for table `tickets`
--

CREATE TABLE `tickets` (
  `id` bigint UNSIGNED NOT NULL,
  `tracking_id` varchar(255) COLLATE utf8mb4_unicode_ci NOT NULL,
  `hari_masuk` date NOT NULL,
  `waktu_masuk` time NOT NULL,
  `hari_respon` date DEFAULT NULL,
  `waktu_respon` time DEFAULT NULL,
  `solved_time` varchar(255) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `user_name` varchar(255) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `user_email` varchar(255) COLLATE utf8mb4_unicode_ci NOT NULL,
  `no_whatsapp` varchar(50) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `category_name` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL,
  `products_name` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci DEFAULT NULL,
  `priority` enum('low','medium','high','critical') COLLATE utf8mb4_unicode_ci NOT NULL,
  `status` enum('New','On Progress','Resolved') COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT 'New',
  `subject` varchar(255) COLLATE utf8mb4_unicode_ci NOT NULL,
  `detail_kendala` text COLLATE utf8mb4_unicode_ci NOT NULL,
  `PIC` varchar(255) COLLATE utf8mb4_unicode_ci NOT NULL,
  `time_worked` int DEFAULT NULL,
  `due_date` date DEFAULT NULL,
  `respon_diberikan` text COLLATE utf8mb4_unicode_ci,
  `created_at` timestamp NULL DEFAULT NULL,
  `updated_at` timestamp NULL DEFAULT NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

--
-- Dumping data for table `tickets`
--

INSERT INTO `tickets` (`id`, `tracking_id`, `hari_masuk`, `waktu_masuk`, `hari_respon`, `waktu_respon`, `solved_time`, `user_name`, `user_email`, `no_whatsapp`, `category_name`, `products_name`, `priority`, `status`, `subject`, `detail_kendala`, `PIC`, `time_worked`, `due_date`, `respon_diberikan`, `created_at`, `updated_at`) VALUES
(73, 'P-250B-200', '2025-02-26', '14:30:00', '2025-02-27', '10:15:00', '0 hours 12 minutes 47 seconds', 'Wildan', 'mwildab15@gmail.com', '6281234567890', 'Kendala Login', 'Photobooth', 'high', 'Resolved', 'Network Connection Issue', 'Unable to connect to company VPN from remote location', 'Jibril', NULL, NULL, 'Checked VPN server logs, identified authentication issue. Provided temporary credentials.', '2025-02-26 08:54:31', '2025-02-26 09:07:18'),
(74, 'GP-250V-828', '2025-02-26', '14:30:00', '2025-02-27', '10:15:00', '0 hours 9 minutes 54 seconds', 'Wildan', 'mwildab15@gmail.com', '6281234567890', 'Kendala Login', 'Gugus Pangan', 'high', 'Resolved', 'Network Connection Issue', 'Unable to connect to company VPN from remote location', 'Jibril', NULL, NULL, 'Checked VPN server logs, identified authentication issue. Provided temporary credentials.', '2025-02-26 08:54:37', '2025-02-26 09:04:42'),
(75, 'GP-250T-022', '2025-02-26', '14:30:00', '2025-02-27', '10:15:00', '0 hours 3 minutes 26 seconds', 'Wildan', 'mwildab15@gmail.com', '6281234567890', 'Kendala Login', 'Gugus Pangan', 'high', 'Resolved', 'Network Connection Issue', 'Unable to connect to company VPN from remote location', 'Jibril', NULL, NULL, 'Checked VPN server logs, identified authentication issue. Provided temporary credentials.', '2025-02-26 08:59:54', '2025-02-26 09:03:37'),
(76, 'P-250D-079', '2025-02-26', '14:30:00', '2025-02-27', '10:15:00', '0 hours 0 minutes 0 seconds', 'Wildan', 'mwildab15@gmail.com', '6281234567890', 'Kendala Login', 'Photobooth', 'high', 'On Progress', 'Network Connection Issue', 'Unable to connect to company VPN from remote location', 'Jibril', NULL, NULL, 'Checked VPN server logs, identified authentication issue. Provided temporary credentials.', '2025-02-26 09:00:04', '2025-02-27 04:11:07'),
(77, 'GP-250Q-303', '2025-02-26', '14:30:00', '2025-02-27', '10:15:00', '0 hours 0 minutes 47 seconds', 'Wildan', 'mwildab15@gmail.com', '6281234567890', 'Kendala Login', 'Gugus Pangan', 'high', 'Resolved', 'Kendala 123', 'Unable to connect to company VPN from remote location', 'Jibril', NULL, NULL, 'Checked VPN server logs, identified authentication issue. Provided temporary credentials.', '2025-02-26 09:53:44', '2025-02-26 09:54:31'),
(78, 'GP-250K-850', '2025-02-26', '14:30:00', '2025-02-27', '10:15:00', '0 hours 26 minutes 10 seconds', 'Wildan', 'mwildab15@gmail.com', '6281234567890', 'Kendala Login', 'Gugus Pangan', 'high', 'Resolved', 'Kendala 123', 'Unable to connect to company VPN from remote location', 'Jibril', NULL, NULL, 'Checked VPN server logs, identified authentication issue. Provided temporary credentials.', '2025-02-27 02:00:53', '2025-02-27 02:27:03'),
(79, 'GP-250M-216', '2025-02-26', '14:30:00', '2025-02-27', '10:15:00', NULL, 'Wildan', 'mwildab15@gmail.com', '6281234567890', 'Kendala Login', 'Gugus Pangan', 'high', 'New', 'Kendala 123', 'Unable to connect to company VPN from remote location', 'Jibril', NULL, NULL, 'Checked VPN server logs, identified authentication issue. Provided temporary credentials.', '2025-02-27 12:47:42', '2025-02-27 12:47:42'),
(80, 'GP-250P-796', '2025-02-26', '14:30:00', '2025-02-27', '10:15:00', NULL, 'Wildan', 'mwildab15@gmail.com', '6281234567890', 'Kendala Login', 'Gugus Pangan', 'high', 'New', 'Kendala 123', 'Unable to connect to company VPN from remote location', 'Jibril', NULL, NULL, 'Checked VPN server logs, identified authentication issue. Provided temporary credentials.', '2025-02-27 12:50:37', '2025-02-27 12:50:37');

-- --------------------------------------------------------

--
-- Table structure for table `users`
--

CREATE TABLE `users` (
  `id` bigint UNSIGNED NOT NULL,
  `name` varchar(255) COLLATE utf8mb4_unicode_ci NOT NULL,
  `email` varchar(255) COLLATE utf8mb4_unicode_ci NOT NULL,
  `password` varchar(255) COLLATE utf8mb4_unicode_ci NOT NULL,
  `avatar` varchar(255) COLLATE utf8mb4_unicode_ci DEFAULT 'default.jpg',
  `role` varchar(20) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT 'pegawai',
  `status` varchar(255) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `OTP` varchar(255) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `token` text COLLATE utf8mb4_unicode_ci,
  `OTP_Active` tinyint(1) NOT NULL DEFAULT '0',
  `email_verified_at` timestamp NULL DEFAULT NULL,
  `remember_token` varchar(100) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `created_at` timestamp NULL DEFAULT NULL,
  `updated_at` timestamp NULL DEFAULT NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

--
-- Dumping data for table `users`
--

INSERT INTO `users` (`id`, `name`, `email`, `password`, `avatar`, `role`, `status`, `OTP`, `token`, `OTP_Active`, `email_verified_at`, `remember_token`, `created_at`, `updated_at`) VALUES
(7, 'Wildan', 'wildan@example.com', 'securepassword123', 'default.jpg', '', 'offline', NULL, NULL, 0, NULL, NULL, '2025-01-30 04:38:46', '2025-01-30 04:38:46'),
(9, 'Wildan', 'wildan1@example.com', 'securepassword123', 'default.jpg', '', 'offline', NULL, NULL, 0, NULL, NULL, '2025-01-30 04:40:23', '2025-01-30 04:40:23'),
(10, 'Wildan', 'john.doe@example.com', 'securepassword123', 'default.jpg', '', 'offline', NULL, NULL, 0, NULL, NULL, '2025-01-31 03:56:45', '2025-02-18 06:09:00'),
(11, 'Wildan', 'john.doe2@example.com', 'securepassword123', 'default.jpg', '', 'offline', NULL, NULL, 0, NULL, NULL, '2025-02-07 02:06:24', '2025-02-07 02:06:24'),
(13, 'Wildan', 'john.doe3@example.com', 'securepassword123', 'default.jpg', 'pegawai', 'offline', NULL, NULL, 0, NULL, NULL, '2025-02-07 02:07:32', '2025-02-07 02:07:32'),
(14, 'Wildan', 'john.doe5@example.com', 'securepassword123', 'default.jpg', 'pegawai', 'offline', NULL, NULL, 0, NULL, NULL, '2025-02-18 02:30:23', '2025-02-18 02:30:23'),
(16, 'Wildan', 'mwildab15@gmail.com', 'securepassword123', 'people19.png', 'admin', 'offline', NULL, NULL, 0, NULL, NULL, '2025-02-18 02:43:13', '2025-02-26 14:40:14'),
(17, 'Yogaa', 'mwildab16@gmail.com', 'securepassword123', 'default.jpg', 'pegawai', 'offline', NULL, NULL, 0, NULL, NULL, '2025-02-20 07:41:45', '2025-02-27 04:27:20');

-- --------------------------------------------------------

--
-- Table structure for table `user_logs`
--

CREATE TABLE `user_logs` (
  `id` bigint NOT NULL,
  `user_email` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `login_time` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `logout_time` timestamp NULL DEFAULT NULL,
  `shift_name` varchar(255) COLLATE utf8mb4_general_ci DEFAULT NULL,
  `OTP` varchar(255) COLLATE utf8mb4_general_ci NOT NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;

--
-- Dumping data for table `user_logs`
--

INSERT INTO `user_logs` (`id`, `user_email`, `login_time`, `logout_time`, `shift_name`, `OTP`) VALUES
(208, 'mwildab15@gmail.com', '2025-02-19 09:20:34', NULL, NULL, ''),
(209, 'mwildab15@gmail.com', '2025-02-19 12:45:54', NULL, NULL, ''),
(210, 'mwildab15@gmail.com', '2025-02-20 06:09:20', NULL, NULL, ''),
(211, 'mwildab15@gmail.com', '2025-02-20 07:34:32', NULL, NULL, ''),
(212, 'mwildab16@gmail.com', '2025-02-20 07:42:06', NULL, NULL, ''),
(213, 'mwildab16@gmail.com', '2025-02-20 07:44:13', NULL, NULL, ''),
(214, 'mwildab16@gmail.com', '2025-02-21 08:55:22', NULL, NULL, ''),
(215, 'mwildab16@gmail.com', '2025-02-25 03:23:29', NULL, NULL, ''),
(216, 'mwildab16@gmail.com', '2025-02-26 14:31:02', NULL, NULL, ''),
(217, 'mwildab15@gmail.com', '2025-02-26 14:40:14', NULL, NULL, ''),
(218, 'mwildab16@gmail.com', '2025-02-27 04:27:20', NULL, NULL, '');

-- --------------------------------------------------------

--
-- Table structure for table `user_tickets`
--

CREATE TABLE `user_tickets` (
  `id` bigint NOT NULL,
  `tickets_id` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL,
  `user_email` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL,
  `current_status` varchar(255) COLLATE utf8mb4_general_ci NOT NULL,
  `new_status` varchar(50) COLLATE utf8mb4_general_ci NOT NULL,
  `update_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;

--
-- Dumping data for table `user_tickets`
--

INSERT INTO `user_tickets` (`id`, `tickets_id`, `user_email`, `current_status`, `new_status`, `update_at`) VALUES
(104, 'GP-250K-850', 'mwildab15@gmail.com', 'New', 'On Progress', '2025-02-27 02:21:28'),
(105, 'GP-250K-850', 'mwildab15@gmail.com', 'On Progress', 'Resolved', '2025-02-27 02:23:38'),
(106, 'GP-250K-850', 'mwildab15@gmail.com', 'Resolved', 'Resolved', '2025-02-27 02:27:03'),
(107, 'P-250D-079', 'mwildab16@gmail.com', 'Resolved', 'On Progress', '2025-02-27 04:11:07'),
(108, 'GP-250P-796', 'mwildab15@gmail.com', 'New', 'New', '2025-02-27 12:50:37');

--
-- Indexes for dumped tables
--

--
-- Indexes for table `category`
--
ALTER TABLE `category`
  ADD PRIMARY KEY (`id`),
  ADD UNIQUE KEY `category_name` (`category_name`);

--
-- Indexes for table `employee_shifts`
--
ALTER TABLE `employee_shifts`
  ADD PRIMARY KEY (`id`),
  ADD KEY `fk_user_email` (`user_email`),
  ADD KEY `fk_shift_id` (`shift_id`);

--
-- Indexes for table `note`
--
ALTER TABLE `note`
  ADD PRIMARY KEY (`id`),
  ADD KEY `user_email` (`user_email`);

--
-- Indexes for table `products`
--
ALTER TABLE `products`
  ADD PRIMARY KEY (`id`),
  ADD UNIQUE KEY `name` (`name`);

--
-- Indexes for table `shifts`
--
ALTER TABLE `shifts`
  ADD PRIMARY KEY (`id`);

--
-- Indexes for table `shift_logs`
--
ALTER TABLE `shift_logs`
  ADD PRIMARY KEY (`id`),
  ADD KEY `fk_shiftlogs_users` (`user_email`);

--
-- Indexes for table `tenants`
--
ALTER TABLE `tenants`
  ADD PRIMARY KEY (`id`),
  ADD UNIQUE KEY `uni_tenants_name` (`name`);

--
-- Indexes for table `tickets`
--
ALTER TABLE `tickets`
  ADD PRIMARY KEY (`id`),
  ADD UNIQUE KEY `tickets_tracking_id_unique` (`tracking_id`),
  ADD KEY `category_id` (`category_name`),
  ADD KEY `user_emails` (`user_email`),
  ADD KEY `fk_products_name` (`products_name`);

--
-- Indexes for table `users`
--
ALTER TABLE `users`
  ADD PRIMARY KEY (`id`),
  ADD UNIQUE KEY `users_email_unique` (`email`);

--
-- Indexes for table `user_logs`
--
ALTER TABLE `user_logs`
  ADD PRIMARY KEY (`id`),
  ADD KEY `fk_users_logs_user` (`user_email`);

--
-- Indexes for table `user_tickets`
--
ALTER TABLE `user_tickets`
  ADD PRIMARY KEY (`id`),
  ADD KEY `fk_users_email` (`user_email`),
  ADD KEY `fk_tickets_id` (`tickets_id`);

--
-- AUTO_INCREMENT for dumped tables
--

--
-- AUTO_INCREMENT for table `category`
--
ALTER TABLE `category`
  MODIFY `id` bigint NOT NULL AUTO_INCREMENT, AUTO_INCREMENT=4;

--
-- AUTO_INCREMENT for table `employee_shifts`
--
ALTER TABLE `employee_shifts`
  MODIFY `id` bigint NOT NULL AUTO_INCREMENT, AUTO_INCREMENT=28;

--
-- AUTO_INCREMENT for table `note`
--
ALTER TABLE `note`
  MODIFY `id` bigint NOT NULL AUTO_INCREMENT, AUTO_INCREMENT=26;

--
-- AUTO_INCREMENT for table `products`
--
ALTER TABLE `products`
  MODIFY `id` bigint NOT NULL AUTO_INCREMENT, AUTO_INCREMENT=3;

--
-- AUTO_INCREMENT for table `shifts`
--
ALTER TABLE `shifts`
  MODIFY `id` bigint NOT NULL AUTO_INCREMENT, AUTO_INCREMENT=4;

--
-- AUTO_INCREMENT for table `shift_logs`
--
ALTER TABLE `shift_logs`
  MODIFY `id` bigint NOT NULL AUTO_INCREMENT, AUTO_INCREMENT=31;

--
-- AUTO_INCREMENT for table `tenants`
--
ALTER TABLE `tenants`
  MODIFY `id` bigint UNSIGNED NOT NULL AUTO_INCREMENT, AUTO_INCREMENT=4;

--
-- AUTO_INCREMENT for table `tickets`
--
ALTER TABLE `tickets`
  MODIFY `id` bigint UNSIGNED NOT NULL AUTO_INCREMENT, AUTO_INCREMENT=81;

--
-- AUTO_INCREMENT for table `users`
--
ALTER TABLE `users`
  MODIFY `id` bigint UNSIGNED NOT NULL AUTO_INCREMENT, AUTO_INCREMENT=18;

--
-- AUTO_INCREMENT for table `user_logs`
--
ALTER TABLE `user_logs`
  MODIFY `id` bigint NOT NULL AUTO_INCREMENT, AUTO_INCREMENT=219;

--
-- AUTO_INCREMENT for table `user_tickets`
--
ALTER TABLE `user_tickets`
  MODIFY `id` bigint NOT NULL AUTO_INCREMENT, AUTO_INCREMENT=109;

--
-- Constraints for dumped tables
--

--
-- Constraints for table `employee_shifts`
--
ALTER TABLE `employee_shifts`
  ADD CONSTRAINT `fk_shift_id` FOREIGN KEY (`shift_id`) REFERENCES `shifts` (`id`) ON DELETE CASCADE ON UPDATE CASCADE,
  ADD CONSTRAINT `fk_user_email` FOREIGN KEY (`user_email`) REFERENCES `users` (`email`) ON DELETE CASCADE ON UPDATE CASCADE;

--
-- Constraints for table `note`
--
ALTER TABLE `note`
  ADD CONSTRAINT `user_email` FOREIGN KEY (`user_email`) REFERENCES `users` (`email`) ON DELETE CASCADE ON UPDATE CASCADE;

--
-- Constraints for table `shift_logs`
--
ALTER TABLE `shift_logs`
  ADD CONSTRAINT `fk_shiftlogs_users` FOREIGN KEY (`user_email`) REFERENCES `users` (`email`) ON DELETE CASCADE;

--
-- Constraints for table `tickets`
--
ALTER TABLE `tickets`
  ADD CONSTRAINT `category_names` FOREIGN KEY (`category_name`) REFERENCES `category` (`category_name`) ON DELETE CASCADE,
  ADD CONSTRAINT `fk_products_name` FOREIGN KEY (`products_name`) REFERENCES `products` (`name`) ON DELETE CASCADE,
  ADD CONSTRAINT `user_emails` FOREIGN KEY (`user_email`) REFERENCES `users` (`email`) ON DELETE CASCADE;

--
-- Constraints for table `user_logs`
--
ALTER TABLE `user_logs`
  ADD CONSTRAINT `fk_users_logs_user` FOREIGN KEY (`user_email`) REFERENCES `users` (`email`) ON DELETE CASCADE ON UPDATE CASCADE;

--
-- Constraints for table `user_tickets`
--
ALTER TABLE `user_tickets`
  ADD CONSTRAINT `fk_tickets_id` FOREIGN KEY (`tickets_id`) REFERENCES `tickets` (`tracking_id`) ON DELETE CASCADE,
  ADD CONSTRAINT `fk_users_email` FOREIGN KEY (`user_email`) REFERENCES `users` (`email`) ON DELETE CASCADE;
COMMIT;

/*!40101 SET CHARACTER_SET_CLIENT=@OLD_CHARACTER_SET_CLIENT */;
/*!40101 SET CHARACTER_SET_RESULTS=@OLD_CHARACTER_SET_RESULTS */;
/*!40101 SET COLLATION_CONNECTION=@OLD_COLLATION_CONNECTION */;
