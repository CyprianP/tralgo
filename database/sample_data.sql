
-- sample data


INSERT INTO providers (last_name, first_name, phone_number, email) VALUES
    ('LastName_test_1', 'FirstName_test_1',  '078 123 45 67', 'LastName_test_1.FirstName_test_1@test.ch'),
    ('LastName_test_2', 'FirstName_test_2', '078 123 45 67', 'LastName_test_2@test.ch');

INSERT INTO courses (course_name, course_description, provider_id) VALUES
    ('Futures Analysis', 'Chart reading.', 1),
    ('Futures Trades', 'Trade execution timing.', 2);





