ALTER TABLE users ADD CONSTRAINT users_email_unique UNIQUE (email);
ALTER TABLE users ADD CONSTRAINT users_name_unique UNIQUE (name);
ALTER TABLE activities ADD CONSTRAINT user_id_fk FOREIGN KEY (user_id) references users(id) ON DELETE CASCADE;
ALTER TABLE answers ADD CONSTRAINT question_id_fk FOREIGN KEY (question_id) references questions(id) ON DELETE CASCADE;