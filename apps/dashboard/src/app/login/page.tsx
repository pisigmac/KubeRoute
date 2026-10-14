"use client";

import { useEffect, useState } from "react";
import Link from "next/link";
import { apiBase, setToken } from "../../lib/api";

export default function Login() {
  const backendUrl = apiBase();
  const [error, setError] = useState<string | null>(null);
  const [busy, setBusy] = useState(false);

  useEffect(() => {
    const params = new URLSearchParams(window.location.search);
    const token = params.get("token");
    if (token) {
      setToken(token);
      window.location.replace("/dashboard");
    }
  }, []);

  async function localDev() {
    setBusy(true);
    try {
      const res = await fetch(`${backendUrl}/auth/dev?format=json`, {
        headers: { Accept: "application/json" },
      });
      const body = await res.json();
      if (!res.ok) throw new Error(body.error || "dev auth failed");
      setToken(body.token);
      window.location.href = "/dashboard";
    } catch (e) {
      setError(e instanceof Error ? e.message : "failed");
      setBusy(false);
    }
  }

  return (
    <main className="min-h-screen bg-[#F5F0EB] flex items-center justify-center p-8 text-[#1A1A1A]">
      <div className="w-full max-w-sm bg-white p-8 border border-gray-200 rounded-2xl">
        <div className="flex justify-center mb-4">
          <div className="w-10 h-10 rounded-xl bg-black text-white flex items-center justify-center font-bold text-xl">K</div>
        </div>
        <h1 className="text-2xl font-bold mb-1 text-center tracking-tight">Kuberoute AI</h1>
        <p className="text-sm text-gray-500 mb-8 text-center">Sign in to the control plane</p>
        {error && <p className="mb-4 text-sm text-red-700">{error}</p>}
        <div className="space-y-3">
          <a href={`${backendUrl}/auth/github`} className="block w-full px-4 py-3 bg-[#1A1A1A] text-white text-sm text-center rounded-xl font-medium">
            Continue with GitHub
          </a>
          <a href={`${backendUrl}/auth/google`} className="block w-full px-4 py-3 border border-gray-300 text-sm text-center rounded-xl font-medium">
            Continue with Google
          </a>
          <button
            onClick={localDev}
            disabled={busy}
            className="w-full px-4 py-3 border border-dashed border-gray-400 text-sm text-center rounded-xl font-medium"
          >
            {busy ? "Signing in…" : "Local dev sign-in"}
          </button>
        </div>
        <div className="mt-8 pt-6 border-t border-gray-100 text-center">
          <Link href="/" className="text-xs text-gray-500 font-medium">← Back</Link>
        </div>
      </div>
    </main>
  );
}
