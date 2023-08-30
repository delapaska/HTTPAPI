CREATE TABLE users(
	id int PRIMARY KEY NOT NULL
)


CREATE TABLE segments(
	id serial primary key not null,
	name text
)


CREATE TABLE users_segments(
	id serial PRIMARY KEY NOT NULL,
	user_id int REFERENCES users(id) ON DELETE CASCADE,
	segment_id int REFERENCES segments(id) ON DELETE CASCADE,
	status bool,
	TTL timestamp
)


CREATE TABLE action(
	id serial primary key not null,
	name text
)


CREATE TABLE history(
	id serial primary key not null,
	action_id int references action(id) ON DELETE CASCADE,
	user_segment_id int references users_segments(id) ON DELETE CASCADE,
	time timestamp
)



CREATE FUNCTION is_segment_exists(segment_name text) RETURNS BOOL
LANGUAGE PLPGSQL
AS $$
BEGIN
	IF EXISTS (SELECT 1 FROM segments WHERE segments.name = segment_name) THEN
		RETURN true;
	ELSE
		RETURN false;
	END IF;
END
$$;

CREATE OR REPLACE FUNCTION is_user_exists(user_id int) RETURNS BOOL
LANGUAGE PLPGSQL
AS $$
BEGIN
	IF EXISTS (SELECT 1 FROM users WHERE users.id = user_id) THEN
		RETURN true;
	ELSE
		RETURN false;
	END IF;
END
$$;



CREATE OR REPLACE FUNCTION is_action_exists(act text) RETURNS BOOL
LANGUAGE PLPGSQL
AS $$
BEGIN
	IF EXISTS (SELECT 1 FROM action WHERE name = act) THEN
		RETURN true;
	ELSE
		RETURN false;
	END IF;
END
$$;

CREATE OR REPLACE FUNCTION is_ttl_exists(us_id int, segment text) RETURNS BOOL
LANGUAGE PLPGSQL
AS $$
BEGIN
	IF EXISTS (SELECT 1 FROM users_segments WHERE ttl IS NOT NULL AND
			   user_id = us_id AND
			   (SELECT id FROM segments WHERE name = segment) = segment_id)
	THEN
		RETURN true;
	ELSE
		RETURN false;
	END IF;
END
$$;

CREATE OR REPLACE FUNCTION is_time_correct(tm text) RETURNS BOOL
LANGUAGE PLPGSQL
AS $$
BEGIN
	IF to_timestamp(tm, 'YYYY-MM-DD HH24:MI:SS') <> to_timestamp('', 'YYYY-MM-DD HH24:MI:SS')
	THEN
		RETURN true;
	ELSE
		RETURN false;
	END IF;
END
$$;

CREATE OR REPLACE FUNCTION get_user_segments(us_id int) RETURNS SETOF TEXT
LANGUAGE PLPGSQL
AS $$
BEGIN
	IF (SELECT is_user_exists(us_id)) THEN
		RETURN QUERY
		(SELECT name
		FROM segments s JOIN users_segments us ON s.id = us.segment_id
		WHERE us.user_id = us_id AND us.status = true);
	END IF;
END
$$;

SELECT * FROM get_user_history(101, '2023-08-28 10:00:00', '2023-08-28 21:00:00')
CREATE OR REPLACE FUNCTION get_user_history(us_id int, start_time timestamp, end_time timestamp)
RETURNS TABLE (user_id int, segment text, act text, date timestamp)
LANGUAGE PLPGSQL
AS $$
BEGIN
	RETURN QUERY SELECT us.user_id, s.name, a.name, h.time
	FROM segments s
		JOIN users_segments us on s.id = us.segment_id
		JOIN history h on h.user_segment_id = us.id
		JOIN action a on a.id = h.action_id
	WHERE h.time >= start_time AND h.time <= end_time AND us_id = us.user_id;
end
$$;  

CREATE PROCEDURE add_user(id int)
LANGUAGE SQL
AS $$
INSERT INTO users VALUES (id)
$$;


CREATE OR REPLACE FUNCTION get_all_users() RETURNS SETOF INT
LANGUAGE PLPGSQL
AS $$
BEGIN
	RETURN QUERY SELECT * FROM users;
END
$$; 

CREATE OR REPLACE FUNCTION get_all_segments() RETURNS SETOF TEXT
LANGUAGE PLPGSQL
AS $$
BEGIN
	RETURN QUERY SELECT name FROM segments;
END
$$;  


CREATE OR REPLACE PROCEDURE add_segment(segment_name text)
LANGUAGE PLPGSQL
AS $$
BEGIN
	IF NOT (SELECT is_segment_exists(segment_name)) THEN
		INSERT INTO segments(name) VALUES (segment_name);
	END IF;
END
$$;


SELECT * FROM segments
CREATE OR REPLACE PROCEDURE upd_segment(old_name text, new_name text)
LANGUAGE PLPGSQL
AS $$
BEGIN
	UPDATE segments SET name = new_name WHERE segments.name = old_name;
END
$$;



CREATE OR REPLACE PROCEDURE del_segment(segment_name text)
LANGUAGE PLPGSQL
AS $$
BEGIN
	IF (SELECT is_segment_exists(segment_name)) THEN
		DELETE FROM segments WHERE segments.name = segment_name;
	END IF;
END
$$;

 
CREATE OR REPLACE PROCEDURE add_user_seg(us_id int, segment_name text, tm text)
LANGUAGE PLPGSQL
AS $$
BEGIN
	IF  (SELECT is_segment_exists(segment_name)) AND
		(SELECT is_user_exists(us_id)) AND
		(SELECT is_time_correct(tm)) AND
		NOT EXISTS (SELECT 1 FROM users_segments WHERE user_id = us_id
				AND segment_id = (SELECT id FROM segments WHERE name = segment_name))	
	THEN
		RAISE NOTICE 'IF COMPLETED';
		INSERT INTO users_segments(user_id, segment_id, status, ttl) VALUES (
			us_id,
			(SELECT id FROM segments WHERE segments.name = segment_name),
			true,
			to_timestamp(tm, 'YYYY-MM-DD HH24:MI:SS')
		);
	ELSIF EXISTS (SELECT 1 FROM users_segments WHERE user_id = us_id
				AND segment_id = (SELECT id FROM segments WHERE name = segment_name))	
	THEN
		RAISE NOTICE 'IF2 COMPLETED';
		UPDATE users_segments SET
			status = false
		WHERE   user_id = us_id AND
				segment_id = (SELECT id FROM segments WHERE name = segment_name);
	ELSIF (SELECT is_segment_exists(segment_name)) AND
		(SELECT is_user_exists(us_id)) AND
		NOT (SELECT is_time_correct(tm))
	THEN
		RAISE NOTICE 'IF3 COMPLETED';
		INSERT INTO users_segments(user_id, segment_id, status, ttl) VALUES (
			us_id,
			(SELECT id FROM segments WHERE segments.name = segment_name),
			true,
			null
		);
	END IF;
END
$$;  

CREATE OR REPLACE PROCEDURE check_ttl(now timestamp)
LANGUAGE PLPGSQL
AS $$
DECLARE
    user_segment_cursor CURSOR FOR
        SELECT id, user_id, segment_id, ttl
        FROM users_segments
        WHERE status = true AND ttl IS NOT NULL;
    user_segment_id int;
    user_id int;
    segment_id int;
    ttl timestamp;
BEGIN
    OPEN user_segment_cursor;
    LOOP
        FETCH user_segment_cursor INTO user_segment_id, user_id, segment_id, ttl;
        EXIT WHEN NOT FOUND;
        
        IF ttl <= now THEN
            CALL del_user_seg_by_id(user_id, segment_id);
			CALL add_to_history_by_id(user_id, segment_id, now, 'Deleted');
        END IF;
    END LOOP;
    CLOSE user_segment_cursor;
END
$$;



CREATE OR REPLACE PROCEDURE del_user_seg_by_id(us_id int, seg_id int)
LANGUAGE PLPGSQL
AS $$
BEGIN
    UPDATE users_segments SET
        status = false
    WHERE   user_id = us_id AND
            segment_id = seg_id;
END
$$;

CREATE OR REPLACE PROCEDURE del_user_seg(us_id int, segment_name text)
LANGUAGE PLPGSQL
AS $$
BEGIN
	UPDATE users_segments SET
			status = false
		WHERE   user_id = us_id AND
				segment_id = (SELECT id FROM segments WHERE name = segment_name);
END
$$;



CREATE OR REPLACE PROCEDURE add_to_history(us_id int, segment text, date timestamp, act text)
LANGUAGE PLPGSQL
AS $$
BEGIN
	IF (SELECT is_user_exists(us_id)) AND
		(SELECT is_segment_exists(segment)) AND
		(SELECT is_action_exists(act)) AND
		EXISTS (SELECT id FROM users_segments WHERE user_id = us_id AND
				segment_id = (SELECT id FROM segments WHERE name = segment))
	THEN
		INSERT INTO history(action_id, user_segment_id, time) VALUES
		((SELECT id FROM action WHERE name = act),
		 (SELECT id FROM users_segments WHERE user_id = us_id AND
				segment_id = (SELECT id FROM segments WHERE name = segment)),
		 date);
	END IF;
END
$$;


CREATE OR REPLACE PROCEDURE add_to_history_by_id(us_id int, seg_id int, date timestamp, act text)
LANGUAGE PLPGSQL
AS $$
BEGIN
	IF EXISTS (SELECT id FROM users_segments WHERE user_id = us_id AND
				segment_id = seg_id)
	THEN
		INSERT INTO history(action_id, user_segment_id, time) VALUES
		((SELECT id FROM action WHERE name = act),
		 (SELECT id FROM users_segments WHERE user_id = us_id AND
				segment_id = seg_id),
		 date);
	END IF;
END
$$;
		

