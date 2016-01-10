
CREATE TABLE "games" (
	"game_id" SERIAL PRIMARY KEY,
	"map_name" TEXT NOT NULL,
	"game_mode" TEXT NOT NULL,
	"minimum_level" INTEGER DEFAULT 0,
	"player_count" INTEGER DEFAULT 0,
	"maximum_players" INTEGER DEFAULT 16
);

CREATE TABLE "account_data" (
	"user_id" SERIAL PRIMARY KEY,
	"username" TEXT NOT NULL,
	"password" BYTEA NOT NULL,
	"salt" BYTEA NOT NULL,
	"algorithm" TEXT NOT NULL,
	"createdon" TIMESTAMP NOT NULL,
	"lastlogin" TIMESTAMP NOT NULL
);

CREATE TABLE "characters" (
	"id" SERIAL PRIMARY KEY,
	"uid" INTEGER references account_data,
	"name" TEXT,
	"game_data" JSON,
	"last_game_id" INTEGER DEFAULT 0
);


CREATE TABLE "machines" (
	"machine_id" SERIAL PRIMARY KEY,
	"remote_address" TEXT,
	"service_listen_port" INTEGER
);

CREATE TABLE "machines_metadata" (
	"machine_id" INTEGER PRIMARY KEY references machines(machine_id) ON DELETE CASCADE,
	"most_recent_key" TEXT,
	"last_heartbeat" TIMESTAMP,
	"cpu_usage_pct" REAL,
	"network_usage_pct" REAL,
	"player_occupancy_pct" REAL
);

CREATE TABLE "loading_hosts" (
	"game_id" SERIAL PRIMARY KEY references games(game_id) DEFERRABLE INITIALLY DEFERRED,
	"machine_id" SERIAL references machines(machine_id),
	"kickoff_time" TIMESTAMP
);

CREATE TABLE "hosts" (
	"game_id" SERIAL PRIMARY KEY references games(game_id),
	"machine_id" SERIAL references machines(machine_id),
	"port" INTEGER
);

CREATE FUNCTION get_available_machine()
	RETURNS TABLE (
		"remote_address" TEXT,
		"service_listen_port" INTEGER,
		"most_recent_key" TEXT
	) AS
$$
BEGIN
	RETURN QUERY
	SELECT m.remote_address, m.service_listen_port, mm.most_recent_key
	FROM machines m
	  JOIN machines_metadata mm USING (machine_id)
	WHERE mm.cpu_usage_pct < 80.0
	AND mm.network_usage_pct < 80.0
	ORDER BY RANDOM()
	LIMIT 1;
END
$$ language plpgsql;
