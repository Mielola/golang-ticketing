-- phpMyAdmin SQL Dump
-- version 5.2.1
-- https://www.phpmyadmin.net/
--
-- Host: 127.0.0.1
-- Waktu pembuatan: 20 Feb 2025 pada 07.03
-- Versi server: 10.4.32-MariaDB
-- Versi PHP: 8.2.12

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
-- Struktur dari tabel `category`
--

CREATE TABLE `category` (
  `id` bigint(20) NOT NULL,
  `category_name` varchar(255) NOT NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;

--
-- Dumping data untuk tabel `category`
--

INSERT INTO `category` (`id`, `category_name`) VALUES
(1, 'Kendala Login');

-- --------------------------------------------------------

--
-- Struktur dari tabel `employee_shifts`
--

CREATE TABLE `employee_shifts` (
  `id` bigint(20) NOT NULL,
  `user_email` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL,
  `shift_id` bigint(20) NOT NULL,
  `shift_date` date NOT NULL,
  `created_at` timestamp NOT NULL DEFAULT current_timestamp()
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;

--
-- Dumping data untuk tabel `employee_shifts`
--

INSERT INTO `employee_shifts` (`id`, `user_email`, `shift_id`, `shift_date`, `created_at`) VALUES
(25, 'john.doe@example.com', 3, '2025-02-01', '2025-01-31 04:04:36'),
(26, 'john.doe@example.com', 2, '2025-01-31', '2025-01-31 04:39:17'),
(27, 'mwildab16@gmail.com', 2, '2025-02-19', '2025-01-31 07:21:36');

-- --------------------------------------------------------

--
-- Struktur dari tabel `note`
--

CREATE TABLE `note` (
  `id` bigint(26) NOT NULL,
  `Title` varchar(256) NOT NULL,
  `Content` varchar(256) NOT NULL,
  `user_email` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci DEFAULT NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;

-- --------------------------------------------------------

--
-- Struktur dari tabel `password_reset_tokens`
--

CREATE TABLE `password_reset_tokens` (
  `email` varchar(255) NOT NULL,
  `token` varchar(255) NOT NULL,
  `created_at` timestamp NULL DEFAULT NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- --------------------------------------------------------

--
-- Struktur dari tabel `personal_access_tokens`
--

CREATE TABLE `personal_access_tokens` (
  `id` bigint(20) UNSIGNED NOT NULL,
  `tokenable_type` varchar(255) NOT NULL,
  `tokenable_id` bigint(20) UNSIGNED NOT NULL,
  `name` varchar(255) NOT NULL,
  `token` varchar(64) NOT NULL,
  `abilities` text DEFAULT NULL,
  `last_used_at` timestamp NULL DEFAULT NULL,
  `expires_at` timestamp NULL DEFAULT NULL,
  `created_at` timestamp NULL DEFAULT NULL,
  `updated_at` timestamp NULL DEFAULT NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- --------------------------------------------------------

--
-- Struktur dari tabel `shifts`
--

CREATE TABLE `shifts` (
  `id` bigint(20) NOT NULL,
  `shift_name` varchar(50) DEFAULT NULL,
  `start_time` time DEFAULT NULL,
  `end_time` time DEFAULT NULL,
  `created_at` timestamp NOT NULL DEFAULT current_timestamp()
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;

--
-- Dumping data untuk tabel `shifts`
--

INSERT INTO `shifts` (`id`, `shift_name`, `start_time`, `end_time`, `created_at`) VALUES
(1, 'Pagi', '08:00:00', '15:00:00', '2025-01-30 03:17:16'),
(2, 'Sore', '15:00:00', '23:00:00', '2025-01-30 03:28:05'),
(3, 'Malam', '23:00:00', '07:00:00', '2025-01-30 03:28:47');

-- --------------------------------------------------------

--
-- Struktur dari tabel `shift_logs`
--

CREATE TABLE `shift_logs` (
  `id` bigint(20) NOT NULL,
  `user_email` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `shift_id` bigint(20) NOT NULL,
  `shift_date` date NOT NULL,
  `reason` text NOT NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;

--
-- Dumping data untuk tabel `shift_logs`
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
-- Struktur dari tabel `tenants`
--

CREATE TABLE `tenants` (
  `id` bigint(20) UNSIGNED NOT NULL,
  `name` varchar(191) DEFAULT NULL,
  `database` longtext DEFAULT NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;

--
-- Dumping data untuk tabel `tenants`
--

INSERT INTO `tenants` (`id`, `name`, `database`) VALUES
(1, 'Tenant1', 'tenant_Tenant1'),
(2, 'Whoops', 'tenant_Whoops'),
(3, 'Whoops1', 'tenants_Whoops1');

-- --------------------------------------------------------

--
-- Struktur dari tabel `tickets`
--

CREATE TABLE `tickets` (
  `id` bigint(20) UNSIGNED NOT NULL,
  `tracking_id` varchar(255) NOT NULL,
  `hari_masuk` date NOT NULL,
  `waktu_masuk` time NOT NULL,
  `hari_respon` date DEFAULT NULL,
  `waktu_respon` time DEFAULT NULL,
  `user_name` varchar(255) DEFAULT NULL,
  `user_email` varchar(255) NOT NULL,
  `category` bigint(20) NOT NULL,
  `priority` enum('low','medium','high','critical') NOT NULL,
  `status` enum('On Progress','New','Resolved') NOT NULL,
  `subject` varchar(255) NOT NULL,
  `detail_kendala` text NOT NULL,
  `owner` varchar(255) NOT NULL,
  `time_worked` int(11) DEFAULT NULL,
  `due_date` date DEFAULT NULL,
  `respon_diberikan` text DEFAULT NULL,
  `created_at` timestamp NULL DEFAULT NULL,
  `updated_at` timestamp NULL DEFAULT NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

--
-- Dumping data untuk tabel `tickets`
--

INSERT INTO `tickets` (`id`, `tracking_id`, `hari_masuk`, `waktu_masuk`, `hari_respon`, `waktu_respon`, `user_name`, `user_email`, `category`, `priority`, `status`, `subject`, `detail_kendala`, `owner`, `time_worked`, `due_date`, `respon_diberikan`, `created_at`, `updated_at`) VALUES
(29, 'TKT-007', '2025-02-06', '08:30:00', NULL, NULL, 'John Doe', 'john.doe@example.com', 1, 'high', 'New', 'Server Down', 'Server utama tidak bisa diakses sejak pagi', 'IT Support', NULL, NULL, NULL, '2025-02-06 03:49:22', '2025-02-06 03:49:22'),
(31, '', '2025-02-06', '08:30:00', NULL, NULL, 'John Doe', 'john.doe@example.com', 1, 'high', 'New', 'Server Down', 'Server utama tidak bisa diakses sejak pagi', 'IT Support', NULL, NULL, NULL, '2025-02-06 03:54:57', '2025-02-06 03:54:57'),
(33, '250-206-L10', '2025-02-06', '00:00:08', '0000-00-00', '00:00:00', 'dudul', 'john.doe@example.com', 1, 'high', 'New', '', '', '', NULL, NULL, '', '2025-02-06 06:10:17', '2025-02-06 06:10:17'),
(34, '250-206-E14', '2025-02-06', '00:00:08', '0001-01-01', '00:00:00', 'dudul', 'john.doe@example.com', 1, 'high', 'Resolved', '', '', '', NULL, NULL, '', '2025-02-06 06:12:14', '2025-02-06 06:58:40'),
(35, '250-206-H38', '2025-02-06', '00:00:08', '0000-00-00', '00:00:00', 'dudul', 'john.doe@example.com', 1, 'high', 'New', 'Masalah Login', 'Tidak bisa login menggunakan nrp', '', NULL, NULL, '', '2025-02-06 06:13:44', '2025-02-06 06:13:44'),
(36, '250-206-U63', '2025-02-06', '00:00:08', '0000-00-00', '00:00:00', 'dudul', 'john.doe@example.com', 1, 'high', 'New', 'Masalah Login', 'Tidak bisa login menggunakan nrp', '', NULL, NULL, '', '2025-02-06 06:56:43', '2025-02-06 06:56:43');

-- --------------------------------------------------------

--
-- Struktur dari tabel `users`
--

CREATE TABLE `users` (
  `id` bigint(20) UNSIGNED NOT NULL,
  `name` varchar(255) NOT NULL,
  `email` varchar(255) NOT NULL,
  `password` varchar(255) NOT NULL,
  `avatar` varchar(255) DEFAULT NULL,
  `role` varchar(20) NOT NULL DEFAULT '''pegawai''',
  `status` varchar(255) DEFAULT NULL,
  `OTP` varchar(255) DEFAULT NULL,
  `token` text DEFAULT NULL,
  `OTP_Active` tinyint(1) NOT NULL DEFAULT 0,
  `email_verified_at` timestamp NULL DEFAULT NULL,
  `remember_token` varchar(100) DEFAULT NULL,
  `created_at` timestamp NULL DEFAULT NULL,
  `updated_at` timestamp NULL DEFAULT NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

--
-- Dumping data untuk tabel `users`
--

INSERT INTO `users` (`id`, `name`, `email`, `password`, `avatar`, `role`, `status`, `OTP`, `token`, `OTP_Active`, `email_verified_at`, `remember_token`, `created_at`, `updated_at`) VALUES
(7, 'Wildan', 'wildan@example.com', 'securepassword123', NULL, '', 'offline', NULL, NULL, 0, NULL, NULL, '2025-01-30 04:38:46', '2025-01-30 04:38:46'),
(9, 'Wildan', 'wildan1@example.com', 'securepassword123', NULL, '', 'offline', NULL, NULL, 0, NULL, NULL, '2025-01-30 04:40:23', '2025-01-30 04:40:23'),
(10, 'Wildan', 'john.doe@example.com', 'securepassword123', NULL, '', 'online', '675990', NULL, 0, NULL, NULL, '2025-01-31 03:56:45', '2025-02-18 06:09:00'),
(11, 'Wildan', 'john.doe2@example.com', 'securepassword123', NULL, '', 'offline', NULL, NULL, 0, NULL, NULL, '2025-02-07 02:06:24', '2025-02-07 02:06:24'),
(13, 'Wildan', 'john.doe3@example.com', 'securepassword123', NULL, '\'pegawai\'', 'offline', NULL, NULL, 0, NULL, NULL, '2025-02-07 02:07:32', '2025-02-07 02:07:32'),
(14, 'Wildan', 'john.doe5@example.com', 'securepassword123', NULL, '\'pegawai\'', 'offline', NULL, NULL, 0, NULL, NULL, '2025-02-18 02:30:23', '2025-02-18 02:30:23'),
(16, 'wildansss', 'mwildab16@gmail.com', 'securepassword123', 'celana.png', 'admin', 'online', NULL, 'eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VybmFtZSI6IjUwMDA0NCJ9.UHl9BNNFyc0VUTJulXj0-nugnmnHVDsMSUaKlqfJ49k', 0, NULL, NULL, '2025-02-18 02:43:13', '2025-02-19 12:45:54');

-- --------------------------------------------------------

--
-- Struktur dari tabel `user_logs`
--

CREATE TABLE `user_logs` (
  `id` bigint(20) NOT NULL,
  `user_email` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  `login_time` timestamp NOT NULL DEFAULT current_timestamp(),
  `logout_time` timestamp NULL DEFAULT NULL,
  `shift_name` varchar(255) DEFAULT NULL,
  `OTP` varchar(255) NOT NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;

--
-- Dumping data untuk tabel `user_logs`
--

INSERT INTO `user_logs` (`id`, `user_email`, `login_time`, `logout_time`, `shift_name`, `OTP`) VALUES
(208, 'mwildab16@gmail.com', '2025-02-19 09:20:34', NULL, NULL, ''),
(209, 'mwildab16@gmail.com', '2025-02-19 12:45:54', NULL, NULL, '');

-- --------------------------------------------------------

--
-- Struktur dari tabel `user_tickets`
--

CREATE TABLE `user_tickets` (
  `id` bigint(20) NOT NULL,
  `tickets_id` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL,
  `user_email` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL,
  `status` varchar(255) NOT NULL,
  `update_at` timestamp NOT NULL DEFAULT current_timestamp() ON UPDATE current_timestamp()
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;

--
-- Dumping data untuk tabel `user_tickets`
--

INSERT INTO `user_tickets` (`id`, `tickets_id`, `user_email`, `status`, `update_at`) VALUES
(11, '250-206-U63', 'john.doe@example.com', 'New', '2025-02-06 06:56:43'),
(12, '250-206-E14', 'john.doe@example.com', 'On Progress', '2025-02-06 06:57:30'),
(13, '250-206-E14', 'john.doe@example.com', 'Resolved', '2025-02-06 06:58:40');

--
-- Indexes for dumped tables
--

--
-- Indeks untuk tabel `category`
--
ALTER TABLE `category`
  ADD PRIMARY KEY (`id`);

--
-- Indeks untuk tabel `employee_shifts`
--
ALTER TABLE `employee_shifts`
  ADD PRIMARY KEY (`id`),
  ADD KEY `fk_user_email` (`user_email`),
  ADD KEY `fk_shift_id` (`shift_id`);

--
-- Indeks untuk tabel `note`
--
ALTER TABLE `note`
  ADD PRIMARY KEY (`id`),
  ADD KEY `user_email` (`user_email`);

--
-- Indeks untuk tabel `password_reset_tokens`
--
ALTER TABLE `password_reset_tokens`
  ADD PRIMARY KEY (`email`);

--
-- Indeks untuk tabel `personal_access_tokens`
--
ALTER TABLE `personal_access_tokens`
  ADD PRIMARY KEY (`id`),
  ADD UNIQUE KEY `personal_access_tokens_token_unique` (`token`),
  ADD KEY `personal_access_tokens_tokenable_type_tokenable_id_index` (`tokenable_type`,`tokenable_id`);

--
-- Indeks untuk tabel `shifts`
--
ALTER TABLE `shifts`
  ADD PRIMARY KEY (`id`);

--
-- Indeks untuk tabel `shift_logs`
--
ALTER TABLE `shift_logs`
  ADD PRIMARY KEY (`id`),
  ADD KEY `fk_shiftlogs_users` (`user_email`);

--
-- Indeks untuk tabel `tenants`
--
ALTER TABLE `tenants`
  ADD PRIMARY KEY (`id`),
  ADD UNIQUE KEY `uni_tenants_name` (`name`);

--
-- Indeks untuk tabel `tickets`
--
ALTER TABLE `tickets`
  ADD PRIMARY KEY (`id`),
  ADD UNIQUE KEY `tickets_tracking_id_unique` (`tracking_id`),
  ADD KEY `category_id` (`category`),
  ADD KEY `user_emails` (`user_email`);

--
-- Indeks untuk tabel `users`
--
ALTER TABLE `users`
  ADD PRIMARY KEY (`id`),
  ADD UNIQUE KEY `users_email_unique` (`email`);

--
-- Indeks untuk tabel `user_logs`
--
ALTER TABLE `user_logs`
  ADD PRIMARY KEY (`id`),
  ADD KEY `fk_users_logs_user` (`user_email`);

--
-- Indeks untuk tabel `user_tickets`
--
ALTER TABLE `user_tickets`
  ADD PRIMARY KEY (`id`),
  ADD KEY `fk_users_email` (`user_email`),
  ADD KEY `fk_tickets_id` (`tickets_id`);

--
-- AUTO_INCREMENT untuk tabel yang dibuang
--

--
-- AUTO_INCREMENT untuk tabel `category`
--
ALTER TABLE `category`
  MODIFY `id` bigint(20) NOT NULL AUTO_INCREMENT, AUTO_INCREMENT=2;

--
-- AUTO_INCREMENT untuk tabel `employee_shifts`
--
ALTER TABLE `employee_shifts`
  MODIFY `id` bigint(20) NOT NULL AUTO_INCREMENT, AUTO_INCREMENT=28;

--
-- AUTO_INCREMENT untuk tabel `note`
--
ALTER TABLE `note`
  MODIFY `id` bigint(26) NOT NULL AUTO_INCREMENT, AUTO_INCREMENT=25;

--
-- AUTO_INCREMENT untuk tabel `personal_access_tokens`
--
ALTER TABLE `personal_access_tokens`
  MODIFY `id` bigint(20) UNSIGNED NOT NULL AUTO_INCREMENT;

--
-- AUTO_INCREMENT untuk tabel `shifts`
--
ALTER TABLE `shifts`
  MODIFY `id` bigint(20) NOT NULL AUTO_INCREMENT, AUTO_INCREMENT=4;

--
-- AUTO_INCREMENT untuk tabel `shift_logs`
--
ALTER TABLE `shift_logs`
  MODIFY `id` bigint(20) NOT NULL AUTO_INCREMENT, AUTO_INCREMENT=31;

--
-- AUTO_INCREMENT untuk tabel `tenants`
--
ALTER TABLE `tenants`
  MODIFY `id` bigint(20) UNSIGNED NOT NULL AUTO_INCREMENT, AUTO_INCREMENT=4;

--
-- AUTO_INCREMENT untuk tabel `tickets`
--
ALTER TABLE `tickets`
  MODIFY `id` bigint(20) UNSIGNED NOT NULL AUTO_INCREMENT, AUTO_INCREMENT=37;

--
-- AUTO_INCREMENT untuk tabel `users`
--
ALTER TABLE `users`
  MODIFY `id` bigint(20) UNSIGNED NOT NULL AUTO_INCREMENT, AUTO_INCREMENT=17;

--
-- AUTO_INCREMENT untuk tabel `user_logs`
--
ALTER TABLE `user_logs`
  MODIFY `id` bigint(20) NOT NULL AUTO_INCREMENT, AUTO_INCREMENT=210;

--
-- AUTO_INCREMENT untuk tabel `user_tickets`
--
ALTER TABLE `user_tickets`
  MODIFY `id` bigint(20) NOT NULL AUTO_INCREMENT, AUTO_INCREMENT=14;

--
-- Ketidakleluasaan untuk tabel pelimpahan (Dumped Tables)
--

--
-- Ketidakleluasaan untuk tabel `employee_shifts`
--
ALTER TABLE `employee_shifts`
  ADD CONSTRAINT `fk_shift_id` FOREIGN KEY (`shift_id`) REFERENCES `shifts` (`id`) ON DELETE CASCADE ON UPDATE CASCADE,
  ADD CONSTRAINT `fk_user_email` FOREIGN KEY (`user_email`) REFERENCES `users` (`email`) ON DELETE CASCADE ON UPDATE CASCADE;

--
-- Ketidakleluasaan untuk tabel `note`
--
ALTER TABLE `note`
  ADD CONSTRAINT `user_email` FOREIGN KEY (`user_email`) REFERENCES `users` (`email`) ON DELETE CASCADE ON UPDATE CASCADE;

--
-- Ketidakleluasaan untuk tabel `shift_logs`
--
ALTER TABLE `shift_logs`
  ADD CONSTRAINT `fk_shiftlogs_users` FOREIGN KEY (`user_email`) REFERENCES `users` (`email`) ON DELETE CASCADE;

--
-- Ketidakleluasaan untuk tabel `tickets`
--
ALTER TABLE `tickets`
  ADD CONSTRAINT `category_id` FOREIGN KEY (`category`) REFERENCES `category` (`id`) ON DELETE CASCADE,
  ADD CONSTRAINT `user_emails` FOREIGN KEY (`user_email`) REFERENCES `users` (`email`) ON DELETE CASCADE;

--
-- Ketidakleluasaan untuk tabel `user_logs`
--
ALTER TABLE `user_logs`
  ADD CONSTRAINT `fk_users_logs_user` FOREIGN KEY (`user_email`) REFERENCES `users` (`email`) ON DELETE CASCADE ON UPDATE CASCADE;

--
-- Ketidakleluasaan untuk tabel `user_tickets`
--
ALTER TABLE `user_tickets`
  ADD CONSTRAINT `fk_tickets_id` FOREIGN KEY (`tickets_id`) REFERENCES `tickets` (`tracking_id`) ON DELETE CASCADE,
  ADD CONSTRAINT `fk_users_email` FOREIGN KEY (`user_email`) REFERENCES `users` (`email`) ON DELETE CASCADE;
COMMIT;

/*!40101 SET CHARACTER_SET_CLIENT=@OLD_CHARACTER_SET_CLIENT */;
/*!40101 SET CHARACTER_SET_RESULTS=@OLD_CHARACTER_SET_RESULTS */;
/*!40101 SET COLLATION_CONNECTION=@OLD_COLLATION_CONNECTION */;
