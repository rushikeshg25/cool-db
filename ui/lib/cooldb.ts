const demoApiUrl = (
  process.env.COOLDB_DEMO_API_URL ?? "http://127.0.0.1:3041"
).replace(/\/$/, "");

export type DemoApiError = {
  error: {
    code: string;
    message: string;
  };
};

export async function requestDemoApi(path: string, init?: RequestInit) {
  return fetch(`${demoApiUrl}${path}`, {
    ...init,
    cache: "no-store",
    signal: AbortSignal.timeout(10_000),
  });
}

export function unavailableResponse() {
  return Response.json(
    {
      error: {
        code: "DEMO_API_UNAVAILABLE",
        message:
          "Start CoolDB with --http-port 3041 before using the local dashboard.",
      },
    } satisfies DemoApiError,
    { status: 503 },
  );
}
