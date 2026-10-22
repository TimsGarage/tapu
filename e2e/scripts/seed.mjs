import crypto from "node:crypto";
import fs from "node:fs/promises";
import { spawnSync } from "node:child_process";
import pg from "pg";

const { Client } = pg;

const iamURL = process.env.E2E_IAM_URL || "http://127.0.0.1:8080";

const receiverURL = process.env.E2E_RECEIVER_URL || "http://127.0.0.1:4100";

const username = process.env.E2E_USERNAME || "e2e-user";

const password = process.env.E2E_PASSWORD;

const adminUsername = process.env.E2E_ADMIN_USERNAME || "e2e-admin";

const adminPassword = process.env.E2E_ADMIN_PASSWORD;

const serviceKey = process.env.E2E_SERVICE_KEY || "e2e-project";

const serviceName = process.env.E2E_SERVICE_NAME || "E2E Project";

const audience = process.env.E2E_AUDIENCE || "e2e-project-api";

const browserClientName = process.env.E2E_CLIENT_NAME || "E2E Browser Client";

const receiverRedirectURI =
	process.env.E2E_REDIRECT_URI || `${receiverURL}/callback`;

const browserClientID = "e2e-browser-client";
const apiClientID = "e2e-api-client";
const apiClientName = "E2E Go API Client";
const apiRedirectURI = "http://127.0.0.1:3000/auth/callback";

if (!password) {
	throw new Error("E2E_PASSWORD is required");
}

if (!adminPassword) {
	throw new Error("E2E_ADMIN_PASSWORD is required");
}

const databaseURL = process.env.E2E_DATABASE_URL;

if (!databaseURL) {
	throw new Error("E2E_DATABASE_URL is required");
}

function hashPassword(value) {
	const result = spawnSync(
		"go",
		["run", "../cmd/iamctl", "password-hash", "--password-stdin"],
		{
			cwd: process.cwd(),
			input: `${value}
`,
			encoding: "utf8",
			env: process.env,
		},
	);

	if (result.status !== 0) {
		throw new Error(
			`Unable to hash password:
${result.stderr}`,
		);
	}

	const hash = result.stdout.trim();

	if (!hash.startsWith("$argon2id$")) {
		throw new Error("iamctl returned an invalid password hash");
	}

	return hash;
}

function secretHash(value) {
	return crypto.createHash("sha256").update(value).digest();
}

async function upsertUser(database, { username, passwordHash, isAdmin }) {
	const proposedID = crypto.randomUUID();

	const result = await database.query(
		`
      INSERT INTO users (
        id,
        username,
        password_hash,
        is_admin,
        status,
        password_changed_at,
        created_at,
        updated_at
      )
      VALUES (
        $1,
        $2,
        $3,
        $4,
        'active',
        NOW(),
        NOW(),
        NOW()
      )
      ON CONFLICT (username) DO UPDATE SET
        password_hash = EXCLUDED.password_hash,
        is_admin = EXCLUDED.is_admin,
        status = 'active',
        password_changed_at = NOW(),
        updated_at = NOW()
      RETURNING id
    `,
		[proposedID, username, passwordHash, isAdmin],
	);

	return result.rows[0].id;
}

async function upsertOAuthClient(
	database,
	{ clientID, name, clientType, secretHash, redirectURI, serviceID },
) {
	const proposedID = crypto.randomUUID();

	const clientResult = await database.query(
		`
      INSERT INTO oauth_clients (
        id,
        client_id,
        name,
        client_type,
        secret_hash,
        enabled,
        created_at,
        updated_at
      )
      VALUES (
        $1,
        $2,
        $3,
        $4,
        $5,
        true,
        NOW(),
        NOW()
      )
      ON CONFLICT (client_id) DO UPDATE SET
        name = EXCLUDED.name,
        client_type = EXCLUDED.client_type,
        secret_hash = EXCLUDED.secret_hash,
        enabled = true,
        updated_at = NOW()
      RETURNING id
    `,
		[proposedID, clientID, name, clientType, secretHash],
	);

	const oauthClientID = clientResult.rows[0].id;

	await database.query(
		`
      INSERT INTO oauth_client_redirect_uris (
        id,
        oauth_client_id,
        redirect_uri
      )
      VALUES ($1, $2, $3)
      ON CONFLICT (
        oauth_client_id,
        redirect_uri
      ) DO NOTHING
    `,
		[crypto.randomUUID(), oauthClientID, redirectURI],
	);

	await database.query(
		`
      INSERT INTO oauth_client_services (
        oauth_client_id,
        service_id,
        created_at
      )
      VALUES ($1, $2, NOW())
      ON CONFLICT (
        oauth_client_id,
        service_id
      ) DO NOTHING
    `,
		[oauthClientID, serviceID],
	);

	return oauthClientID;
}

const userPasswordHash = hashPassword(password);
const adminPasswordHash = hashPassword(adminPassword);

const apiClientSecret = crypto.randomBytes(32).toString("base64url");

const database = new Client({
	connectionString: databaseURL,
});

await database.connect();

try {
	await database.query("BEGIN");

	const userID = await upsertUser(database, {
		username,
		passwordHash: userPasswordHash,
		isAdmin: false,
	});

	await upsertUser(database, {
		username: adminUsername,
		passwordHash: adminPasswordHash,
		isAdmin: true,
	});

	const serviceResult = await database.query(
		`
      INSERT INTO services (
        id,
        key,
        name,
        audience,
        enabled,
        created_at,
        updated_at
      )
      VALUES (
        $1,
        $2,
        $3,
        $4,
        true,
        NOW(),
        NOW()
      )
      ON CONFLICT (key) DO UPDATE SET
        name = EXCLUDED.name,
        audience = EXCLUDED.audience,
        enabled = true,
        updated_at = NOW()
      RETURNING id
    `,
		[crypto.randomUUID(), serviceKey, serviceName, audience],
	);

	const serviceID = serviceResult.rows[0].id;

	await database.query(
		`
      INSERT INTO user_service_access (
        user_id,
        service_id,
        created_at
      )
      VALUES ($1, $2, NOW())
      ON CONFLICT (
        user_id,
        service_id
      ) DO NOTHING
    `,
		[userID, serviceID],
	);

	await upsertOAuthClient(database, {
		clientID: browserClientID,
		name: browserClientName,
		clientType: "public",
		secretHash: null,
		redirectURI: receiverRedirectURI,
		serviceID,
	});

	await upsertOAuthClient(database, {
		clientID: apiClientID,
		name: apiClientName,
		clientType: "confidential",
		secretHash: secretHash(apiClientSecret),
		redirectURI: apiRedirectURI,
		serviceID,
	});

	await database.query("COMMIT");

	await fs.writeFile(
		".generated.json",
		JSON.stringify(
			{
				iamURL,
				receiverURL,
				username,
				password,
				adminUsername,
				adminPassword,
				clientID: browserClientID,
				audience,
				redirectURI: receiverRedirectURI,
				apiClientID,
				apiClientSecret,
				apiRedirectURI,
			},
			null,
			2,
		) + "\n",
		{
			mode: 0o600,
		},
	);

	console.log("E2E seed completed.");
	console.log(`User: ${username}`);
	console.log(`Administrator: ${adminUsername}`);
	console.log(`Browser client: ${browserClientID}`);
	console.log(`API client: ${apiClientID}`);
} catch (error) {
	await database.query("ROLLBACK");
	throw error;
} finally {
	await database.end();
}
