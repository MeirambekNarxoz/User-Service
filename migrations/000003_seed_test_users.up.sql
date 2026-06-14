-- 000003_seed_test_users.up.sql

INSERT INTO users (id, email, password_hash, firstname, lastname, bio, role, status) VALUES
(1, 'admin@example.com',     '$2b$10$EyjYd35d.E8cWRD7FICHeu1xh4r6tAls.QAMc9nRmGgw2kWiDmfZ6', 'Иван',   'Админов',    'Администратор системы',                    'ADMIN',     'ACTIVE'),
(2, 'moderator@example.com', '$2b$10$EyjYd35d.E8cWRD7FICHeu1xh4r6tAls.QAMc9nRmGgw2kWiDmfZ6', 'Петр',   'Модераторов','Модератор споров и апелляций',              'MODERATOR', 'ACTIVE'),
(3, 'expert@example.com',    '$2b$10$EyjYd35d.E8cWRD7FICHeu1xh4r6tAls.QAMc9nRmGgw2kWiDmfZ6', 'Алексей','Экспертов',  'Эксперт по Frontend и Backend разработке', 'USER',      'ACTIVE'),
(4, 'student1@example.com',  '$2b$10$EyjYd35d.E8cWRD7FICHeu1xh4r6tAls.QAMc9nRmGgw2kWiDmfZ6', 'Сергей', 'Студентов',  'Студент направления Frontend',              'USER',      'ACTIVE'),
(5, 'student2@example.com',  '$2b$10$EyjYd35d.E8cWRD7FICHeu1xh4r6tAls.QAMc9nRmGgw2kWiDmfZ6', 'Дмитрий','Учеников',   'Студент направления Frontend',              'USER',      'ACTIVE'),
(6, 'student3@example.com',  '$2b$10$EyjYd35d.E8cWRD7FICHeu1xh4r6tAls.QAMc9nRmGgw2kWiDmfZ6', 'Ольга',  'Проверяющая','Студент направления Backend',               'USER',      'ACTIVE'),
(7, 'student4@example.com',  '$2b$10$EyjYd35d.E8cWRD7FICHeu1xh4r6tAls.QAMc9nRmGgw2kWiDmfZ6', 'Елена',  'Рецензент',  'Студент направления Backend',               'USER',      'ACTIVE')
ON CONFLICT (email) DO UPDATE SET 
    password_hash = EXCLUDED.password_hash,
    firstname = EXCLUDED.firstname,
    lastname = EXCLUDED.lastname,
    role = EXCLUDED.role;

-- Корректируем счетчик sequence для PK id, чтобы не сломать регистрацию новых пользователей
SELECT setval(pg_get_serial_sequence('users', 'id'), coalesce(max(id), 1)) FROM users;
