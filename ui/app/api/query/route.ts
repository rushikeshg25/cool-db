import { requestDemoApi, unavailableResponse } from "@/lib/cooldb";

export async function POST(request: Request) {
  let query: string;
  try {
    const payload: unknown = await request.json();
    if (
      typeof payload !== "object" ||
      payload === null ||
      !("query" in payload) ||
      typeof payload.query !== "string" ||
      payload.query.trim() === ""
    ) {
      throw new Error("invalid query");
    }
    query = payload.query.trim();
  } catch {
    return Response.json(
      {
        error: {
          code: "INVALID_REQUEST",
          message: "Provide a non-empty SQL query.",
        },
      },
      { status: 400 },
    );
  }

  try {
    const response = await requestDemoApi("/api/query", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ query }),
    });
    const payload = await response.json();
    return Response.json(payload, { status: response.status });
  } catch {
    return unavailableResponse();
  }
}
