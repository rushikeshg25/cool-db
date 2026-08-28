import { requestDemoApi, unavailableResponse } from "@/lib/cooldb";

export async function GET() {
  try {
    const response = await requestDemoApi("/api/health");
    const payload = await response.json();
    return Response.json(payload, { status: response.status });
  } catch {
    return unavailableResponse();
  }
}
