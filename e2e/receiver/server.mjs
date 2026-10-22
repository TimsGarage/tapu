import crypto from "node:crypto";
import http from "node:http";
import { URL } from "node:url";

const iamURL = process.env.E2E_IAM_URL || "http://127.0.0.1:8080";

const receiverURL = process.env.E2E_RECEIVER_URL || "http://127.0.0.1:4100";

const clientID = process.env.E2E_CLIENT_ID;
const audience = process.env.E2E_AUDIENCE || "e2e-project-api";

if (!clientID) {
	throw new Error("E2E_CLIENT_ID is required");
}

const transactions = new Map();

function randomValue() {
	return crypto.randomBytes(32).toString("base64url");
}

function pkceChallenge(verifier) {
	return crypto.createHash("sha256").update(verifier).digest("base64url");
}

function html(title, body) {
	return `<!doctype html>
<html lang="en">
<head>
  <meta charset="utf-8">
  <title>${title}</title>
</head>
<body>
  ${body}
</body>
</html>`;
}

const server = http.createServer(async (request, response) => {
	const requestURL = new URL(request.url, receiverURL);

	if (requestURL.pathname === "/") {
		response.writeHead(200, {
			"content-type": "text/html; charset=utf-8",
		});

		response.end(html("E2E receiver", '<a href="/login">Sign in</a>'));
		return;
	}

	if (requestURL.pathname === "/login") {
		const state = randomValue();
		const nonce = randomValue();
		const verifier = randomValue();
		const transactionID = randomValue();

		transactions.set(transactionID, {
			state,
			nonce,
			verifier,
			expiresAt: Date.now() + 10 * 60 * 1000,
		});

		const authorizationURL = new URL("/oauth/authorize", iamURL);

		authorizationURL.search = new URLSearchParams({
			response_type: "code",
			client_id: clientID,
			redirect_uri: `${receiverURL}/callback`,
			scope: "openid profile",
			audience,
			state,
			nonce,
			code_challenge: pkceChallenge(verifier),
			code_challenge_method: "S256",
			transaction_id: transactionID,
		});

		response.writeHead(302, {
			location: authorizationURL.toString(),
		});
		response.end();
		return;
	}

	if (requestURL.pathname === "/callback") {
		const code = requestURL.searchParams.get("code");
		const state = requestURL.searchParams.get("state");

		if (!code || !state) {
			response.writeHead(400, {
				"content-type": "text/html; charset=utf-8",
			});
			response.end(
				html("Callback error", "<p>Missing callback values.</p>"),
			);
			return;
		}

		/*
		 * This receiver stores the transaction ID in the state mapping.
		 * In a production application this belongs in a secure server-side
		 * transaction store or encrypted HttpOnly cookie.
		 */
		const transaction = [...transactions.values()].find(
			(value) => value.state === state,
		);

		if (!transaction || transaction.expiresAt <= Date.now()) {
			response.writeHead(400, {
				"content-type": "text/html; charset=utf-8",
			});
			response.end(html("Callback error", "<p>Invalid state.</p>"));
			return;
		}

		if (transaction.state !== state) {
			response.writeHead(400, {
				"content-type": "text/html; charset=utf-8",
			});
			response.end(html("Callback error", "<p>State mismatch.</p>"));
			return;
		}

		const tokenResponse = await fetch(new URL("/oauth/token", iamURL), {
			method: "POST",
			headers: {
				"content-type": "application/x-www-form-urlencoded",
			},
			body: new URLSearchParams({
				grant_type: "authorization_code",
				client_id: clientID,
				code,
				redirect_uri: `${receiverURL}/callback`,
				code_verifier: transaction.verifier,
			}),
		});

		const tokenBody = await tokenResponse.json();

		if (!tokenResponse.ok) {
			response.writeHead(400, {
				"content-type": "application/json",
			});
			response.end(JSON.stringify(tokenBody));
			return;
		}

		transactions.clear();

		response.writeHead(200, {
			"content-type": "text/html; charset=utf-8",
		});

		response.end(
			html(
				"Authentication successful",
				`<h1>Authentication successful</h1>
         <p id="authenticated">true</p>
         <p id="id-token">${tokenBody.id_token}</p>
         <p id="access-token">${tokenBody.access_token}</p>
         <p id="scope">${tokenBody.scope}</p>`,
			),
		);
		return;
	}

	response.writeHead(404);
	response.end("Not found");
});

server.listen(4100, "127.0.0.1", () => {
	console.log("E2E receiver listening on http://127.0.0.1:4100");
});
