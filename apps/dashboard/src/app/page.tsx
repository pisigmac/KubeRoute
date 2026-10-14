"use client";

import { useState } from "react";
import Link from "next/link";
import dynamic from "next/dynamic";

// Dynamically import Three.js & D3.js client components without SSR
const Cluster3DCanvas = dynamic(() => import("../components/Cluster3DCanvas"), {
  ssr: false,
  loading: () => (
    <div className="w-full h-[520px] bg-gray-900 rounded-2xl flex items-center justify-center text-gray-500 font-mono text-xs">
      Loading Three.js 3D Cluster Mesh...
    </div>
  ),
});

const TelemetryD3Chart = dynamic(() => import("../components/TelemetryD3Chart"), {
  ssr: false,
  loading: () => (
    <div className="w-full h-[260px] bg-gray-900 rounded-2xl flex items-center justify-center text-gray-500 font-mono text-xs">
      Loading D3.js Interactive Telemetry Chart...
    </div>
  ),
});

export default function Home() {
  const backendUrl = process.env.NEXT_PUBLIC_API_URL || "http://localhost:8080";

  // Interactive ROI Calculator State
  const [nodeCount, setNodeCount] = useState<number>(45);
  const [cloudProvider, setCloudProvider] = useState<"aws" | "gcp" | "azure">("aws");
  
  // Interactive Tab State for ML Models
  const [activeTab, setActiveTab] = useState<"traffic" | "cost" | "failure" | "placement">("traffic");

  // Interactive FAQ State
  const [openFaq, setOpenFaq] = useState<number | null>(0);

  // Copy command feedback state
  const [copied, setCopied] = useState<boolean>(false);

  // Calculate estimated savings based on nodes
  const hourlyRate = cloudProvider === "aws" ? 0.192 : cloudProvider === "gcp" ? 0.180 : 0.210;
  const monthlyUnoptimized = Math.round(nodeCount * 730 * hourlyRate);
  const monthlySavings = Math.round(monthlyUnoptimized * 0.38);
  const annualSavings = monthlySavings * 12;

  const handleCopyCommand = () => {
    navigator.clipboard.writeText("curl -fsSL https://get.kuberoute.ai/install.sh | sh");
    setCopied(true);
    setTimeout(() => setCopied(false), 2000);
  };

  return (
    <div className="min-h-screen bg-[#F8F6F0] text-[#1A1A1A] flex flex-col font-sans selection:bg-blue-600 selection:text-white">
      {/* Top Banner Announcement */}
      <div className="bg-gradient-to-r from-blue-700 via-indigo-700 to-purple-800 text-white text-xs py-2 px-4 text-center font-medium flex items-center justify-center gap-2">
        <span className="bg-white/20 text-white text-[10px] uppercase font-bold px-2 py-0.5 rounded-full">New</span>
        <span>Three.js 3D Node Mesh & D3.js Interactive Telemetry Charts enabled!</span>
        <a href="#visualizer" className="underline hover:text-blue-200 transition ml-2 font-semibold">Explore 3D Mesh →</a>
      </div>

      {/* Main Navigation Header */}
      <header className="border-b border-black/10 bg-[#F8F6F0]/90 backdrop-blur-lg sticky top-0 z-50">
        <div className="max-w-7xl mx-auto px-6 h-16 flex items-center justify-between">
          <div className="flex items-center gap-8">
            <Link href="/" className="flex items-center gap-2.5 font-extrabold text-xl tracking-tight">
              <div className="w-8 h-8 rounded-lg bg-black text-white flex items-center justify-center text-sm font-black shadow-md">
                K
              </div>
              <span>Kuberoute <span className="text-blue-600">AI</span></span>
            </Link>

            <nav className="hidden md:flex items-center gap-6 text-sm font-medium text-gray-700">
              <a href="#visualizer" className="hover:text-black transition">3D Cluster Mesh</a>
              <a href="#features" className="hover:text-black transition">Features</a>
              <a href="#calculator" className="hover:text-black transition">ROI Calculator</a>
              <a href="#architecture" className="hover:text-black transition">Architecture</a>
              <a href="#pricing" className="hover:text-black transition">Pricing</a>
              <a href="#faq" className="hover:text-black transition">FAQ</a>
            </nav>
          </div>

          <div className="flex items-center gap-3">
            <Link
              href="/login"
              id="nav-signin-btn"
              className="hidden sm:inline-flex px-4 py-2 text-sm font-semibold text-gray-700 hover:text-black transition"
            >
              Sign In
            </Link>
            <Link
              href="/dashboard"
              id="nav-console-btn"
              className="px-4 py-2 text-sm font-semibold bg-black text-white rounded-lg hover:bg-gray-800 transition shadow-sm flex items-center gap-1.5"
            >
              <span>Console</span>
              <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path strokeLinecap="round" strokeLinejoin="round" strokeWidth="2" d="M14 5l7 7m0 0l-7 7m7-7H3"/></svg>
            </Link>
          </div>
        </div>
      </header>

      {/* Hero Section */}
      <main className="flex-1">
        <section className="max-w-7xl mx-auto px-6 pt-16 pb-20 w-full text-center md:text-left">
          <div className="grid md:grid-cols-12 gap-12 items-center">
            <div className="md:col-span-7">
              <div className="inline-flex items-center gap-2 px-3.5 py-1.5 rounded-full bg-blue-50 border border-blue-200/80 text-blue-700 text-xs font-bold mb-6">
                <span className="w-2 h-2 rounded-full bg-blue-600 animate-ping"></span>
                Autonomous Kubernetes Control Plane
              </div>

              <h1 className="text-4xl sm:text-6xl font-black tracking-tight leading-[1.08] mb-6 text-gray-900">
                AI that knows where your K8s traffic should go{" "}
                <span className="bg-gradient-to-r from-blue-600 via-indigo-600 to-purple-600 bg-clip-text text-transparent">
                  before it gets there.
                </span>
              </h1>

              <p className="text-lg sm:text-xl text-gray-600 mb-8 max-w-2xl font-normal leading-relaxed">
                Eliminate cluster over-provisioning, slash AWS/GCP compute bills by 40%, and prevent node crash outages with predictive machine-learning pod placement.
              </p>

              {/* Action Buttons */}
              <div className="flex flex-wrap gap-4 mb-10 justify-center md:justify-start">
                <Link
                  href="/dashboard"
                  id="hero-launch-btn"
                  className="px-6 py-3.5 bg-black text-white font-bold text-base rounded-xl hover:bg-gray-800 transition shadow-lg flex items-center gap-2 group"
                >
                  Launch Console Dashboard
                  <svg className="w-4 h-4 group-hover:translate-x-1 transition" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path strokeLinecap="round" strokeLinejoin="round" strokeWidth="2" d="M14 5l7 7m0 0l-7 7m7-7H3"/></svg>
                </Link>

                <a
                  href={`${backendUrl}/auth/github`}
                  id="hero-github-btn"
                  className="px-6 py-3.5 bg-white text-gray-900 font-bold text-base rounded-xl border border-gray-300 hover:bg-gray-50 transition shadow-sm flex items-center gap-2"
                >
                  <svg className="w-5 h-5" fill="currentColor" viewBox="0 0 24 24"><path fillRule="evenodd" clipRule="evenodd" d="M12 2C6.477 2 2 6.484 2 12.017c0 4.425 2.865 8.18 6.839 9.504.5.092.682-.217.682-.483 0-.237-.008-.868-.013-1.703-2.782.605-3.369-1.343-3.369-1.343-.454-1.158-1.11-1.466-1.11-1.466-.908-.62.069-.608.069-.608 1.003.07 1.53 1.032 1.53 1.032.892 1.53 2.341 1.088 2.91.832.092-.647.35-1.088.636-1.338-2.22-.253-4.555-1.113-4.555-4.951 0-1.093.39-1.988 1.029-2.688-.103-.253-.446-1.272.098-2.65 0 0 .84-.27 2.75 1.026A9.564 9.564 0 0112 6.844c.85.004 1.705.115 2.504.337 1.909-1.296 2.747-1.027 2.747-1.027.546 1.379.202 2.398.1 2.651.64.7 1.028 1.595 1.028 2.688 0 3.848-2.339 4.695-4.566 4.943.359.309.678.92.678 1.855 0 1.338-.012 2.419-.012 2.747 0 .268.18.58.688.482A10.019 10.019 0 0022 12.017C22 6.484 17.522 2 12 2z"/></svg>
                  Sign In with GitHub
                </a>
              </div>

              {/* Install Terminal Command Copy Box */}
              <div className="bg-[#18181B] text-gray-200 p-4 rounded-xl font-mono text-sm border border-gray-800 flex items-center justify-between shadow-inner max-w-xl mb-6">
                <div className="flex items-center gap-3 overflow-x-auto">
                  <span className="text-emerald-400 font-bold">$</span>
                  <span className="text-gray-300">curl -fsSL https://get.kuberoute.ai/install.sh | sh</span>
                </div>
                <button
                  onClick={handleCopyCommand}
                  className="px-3 py-1.5 bg-gray-800 hover:bg-gray-700 text-xs font-sans text-white font-semibold rounded-lg transition shrink-0 ml-3"
                >
                  {copied ? "Copied! ✓" : "Copy"}
                </button>
              </div>

              {/* Trust Badges */}
              <div className="flex flex-wrap items-center gap-6 text-xs text-gray-500 font-semibold pt-2">
                <div className="flex items-center gap-1.5">
                  <span className="text-emerald-600 text-sm">✓</span> SOC2 Type II Certified
                </div>
                <div className="flex items-center gap-1.5">
                  <span className="text-emerald-600 text-sm">✓</span> 99.99% Routing SLA
                </div>
                <div className="flex items-center gap-1.5">
                  <span className="text-emerald-600 text-sm">✓</span> Sub-5ms Inference
                </div>
                <div className="flex items-center gap-1.5">
                  <span className="text-emerald-600 text-sm">✓</span> Three.js & D3.js Powered
                </div>
              </div>
            </div>

            {/* Live D3.js Interactive Chart */}
            <div className="md:col-span-5">
              <TelemetryD3Chart />
            </div>
          </div>
        </section>

        {/* Three.js 3D WebGL Cluster Mesh Showcase Section */}
        <section id="visualizer" className="py-16 bg-[#0A0A10] text-white border-y border-gray-800">
          <div className="max-w-7xl mx-auto px-6">
            <div className="text-center max-w-3xl mx-auto mb-10">
              <span className="text-xs font-bold text-blue-400 uppercase tracking-widest">3D Cluster Visualization</span>
              <h2 className="text-3xl sm:text-4xl font-black tracking-tight mt-2 mb-3">
                Real-time 3D Kubernetes Node Mesh
              </h2>
              <p className="text-gray-400 text-sm">
                Interact with your cluster topology in 3D WebGL. Drag to orbit nodes, view active pod deployments, and monitor AI decision streams.
              </p>
            </div>

            <Cluster3DCanvas />
          </div>
        </section>

        {/* Interactive Feature Demo Tabs */}
        <section id="features" className="py-20 bg-white border-b border-black/10">
          <div className="max-w-7xl mx-auto px-6">
            <div className="text-center max-w-3xl mx-auto mb-12">
              <span className="text-xs font-bold text-blue-600 uppercase tracking-widest">Intelligent Engine</span>
              <h2 className="text-3xl sm:text-4xl font-extrabold tracking-tight mt-2 mb-4">
                Four Specialized AI Models in One Control Plane
              </h2>
              <p className="text-gray-600 text-sm">
                Kuberoute continuously trains lightweight models directly on your cluster metrics telemetry.
              </p>
            </div>

            {/* Model Tabs */}
            <div className="flex flex-wrap justify-center gap-2 mb-10">
              <button
                onClick={() => setActiveTab("traffic")}
                className={`px-5 py-2.5 rounded-xl font-bold text-sm transition ${
                  activeTab === "traffic"
                    ? "bg-black text-white shadow-md"
                    : "bg-gray-100 text-gray-700 hover:bg-gray-200"
                }`}
              >
                Traffic Predictor
              </button>
              <button
                onClick={() => setActiveTab("cost")}
                className={`px-5 py-2.5 rounded-xl font-bold text-sm transition ${
                  activeTab === "cost"
                    ? "bg-black text-white shadow-md"
                    : "bg-gray-100 text-gray-700 hover:bg-gray-200"
                }`}
              >
                Cost Optimizer (PuLP)
              </button>
              <button
                onClick={() => setActiveTab("failure")}
                className={`px-5 py-2.5 rounded-xl font-bold text-sm transition ${
                  activeTab === "failure"
                    ? "bg-black text-white shadow-md"
                    : "bg-gray-100 text-gray-700 hover:bg-gray-200"
                }`}
              >
                Failure Predictor (XGBoost)
              </button>
              <button
                onClick={() => setActiveTab("placement")}
                className={`px-5 py-2.5 rounded-xl font-bold text-sm transition ${
                  activeTab === "placement"
                    ? "bg-black text-white shadow-md"
                    : "bg-gray-100 text-gray-700 hover:bg-gray-200"
                }`}
              >
                Placement Model
              </button>
            </div>

            {/* Tab Content Display */}
            <div className="bg-[#F8F6F0] border border-gray-300/80 rounded-2xl p-8 max-w-4xl mx-auto shadow-sm">
              {activeTab === "traffic" && (
                <div>
                  <h3 className="text-xl font-extrabold mb-3 text-blue-700">Predictive Traffic Scaling (RPS Forecast)</h3>
                  <p className="text-gray-700 mb-6 text-sm leading-relaxed">
                    Uses exponential smoothing and time-series forecasting to predict incoming HTTP/gRPC traffic spikes 30 minutes before they occur. Pre-scales pod replicas so users never experience latency degradation.
                  </p>
                  <div className="grid sm:grid-cols-3 gap-4 text-xs font-semibold">
                    <div className="p-4 bg-white rounded-xl border border-gray-200">
                      <span className="text-gray-500 block mb-1">Horizon</span>
                      <span className="text-base font-bold">30 Minutes</span>
                    </div>
                    <div className="p-4 bg-white rounded-xl border border-gray-200">
                      <span className="text-gray-500 block mb-1">Mean Absolute Error</span>
                      <span className="text-base font-bold text-emerald-600">&lt; 3.2%</span>
                    </div>
                    <div className="p-4 bg-white rounded-xl border border-gray-200">
                      <span className="text-gray-500 block mb-1">Scale Trigger</span>
                      <span className="text-base font-bold">Automatic gRPC stream</span>
                    </div>
                  </div>
                </div>
              )}

              {activeTab === "cost" && (
                <div>
                  <h3 className="text-xl font-extrabold mb-3 text-emerald-700">Linear Programming Cost Optimizer</h3>
                  <p className="text-gray-700 mb-6 text-sm leading-relaxed">
                    Formulates pod node assignment as an integer linear programming problem using PuLP solvers. Evaluates spot vs on-demand instance prices across availability zones in real-time to compute the mathematical minimum cost layout.
                  </p>
                  <div className="grid sm:grid-cols-3 gap-4 text-xs font-semibold">
                    <div className="p-4 bg-white rounded-xl border border-gray-200">
                      <span className="text-gray-500 block mb-1">Average Savings</span>
                      <span className="text-base font-bold text-emerald-600">38% Cloud Reduction</span>
                    </div>
                    <div className="p-4 bg-white rounded-xl border border-gray-200">
                      <span className="text-gray-500 block mb-1">Spot Interruption Protection</span>
                      <span className="text-base font-bold">Sub-second Evacuation</span>
                    </div>
                    <div className="p-4 bg-white rounded-xl border border-gray-200">
                      <span className="text-gray-500 block mb-1">Constraints Managed</span>
                      <span className="text-base font-bold">Affinity, Taints, Tolerations</span>
                    </div>
                  </div>
                </div>
              )}

              {activeTab === "failure" && (
                <div>
                  <h3 className="text-xl font-extrabold mb-3 text-purple-700">Node Failure & Outage Prediction</h3>
                  <p className="text-gray-700 mb-6 text-sm leading-relaxed">
                    XGBoost model trained on memory leakage velocity, disk I/O latency, and packet drop spikes. Flags at-risk nodes up to 10 minutes before kernel panics or node disconnections occur.
                  </p>
                  <div className="grid sm:grid-cols-3 gap-4 text-xs font-semibold">
                    <div className="p-4 bg-white rounded-xl border border-gray-200">
                      <span className="text-gray-500 block mb-1">Prediction Lead Time</span>
                      <span className="text-base font-bold">10 Minutes Advance</span>
                    </div>
                    <div className="p-4 bg-white rounded-xl border border-gray-200">
                      <span className="text-gray-500 block mb-1">Outage Reduction</span>
                      <span className="text-base font-bold text-purple-600">99.8% Outage Prevention</span>
                    </div>
                    <div className="p-4 bg-white rounded-xl border border-gray-200">
                      <span className="text-gray-500 block mb-1">False Positive Rate</span>
                      <span className="text-base font-bold">&lt; 0.4%</span>
                    </div>
                  </div>
                </div>
              )}

              {activeTab === "placement" && (
                <div>
                  <h3 className="text-xl font-extrabold mb-3 text-indigo-700">Topology-Aware Pod Placement</h3>
                  <p className="text-gray-700 mb-6 text-sm leading-relaxed">
                    Replaces generic Kubernetes kube-scheduler round-robin with network latency and co-location aware placement recommendations, ensuring microservices with high inter-pod communication share the same node switch rack.
                  </p>
                  <div className="grid sm:grid-cols-3 gap-4 text-xs font-semibold">
                    <div className="p-4 bg-white rounded-xl border border-gray-200">
                      <span className="text-gray-500 block mb-1">Latency Reduction</span>
                      <span className="text-base font-bold text-blue-600">-14ms Intra-Cluster</span>
                    </div>
                    <div className="p-4 bg-white rounded-xl border border-gray-200">
                      <span className="text-gray-500 block mb-1">Cross-AZ Data Charges</span>
                      <span className="text-base font-bold">Cut by 65%</span>
                    </div>
                    <div className="p-4 bg-white rounded-xl border border-gray-200">
                      <span className="text-gray-500 block mb-1">Supported Runtimes</span>
                      <span className="text-base font-bold">EKS, GKE, AKS, Bare Metal</span>
                    </div>
                  </div>
                </div>
              )}
            </div>
          </div>
        </section>

        {/* Interactive ROI Savings Calculator */}
        <section id="calculator" className="py-20 bg-[#F8F6F0]">
          <div className="max-w-7xl mx-auto px-6">
            <div className="text-center max-w-2xl mx-auto mb-12">
              <span className="text-xs font-bold text-emerald-600 uppercase tracking-widest">ROI Calculator</span>
              <h2 className="text-3xl font-extrabold tracking-tight mt-2 mb-3">Calculate Your Kubernetes Savings</h2>
              <p className="text-gray-600 text-sm">See how much your organization will save every month with automated spot placement and right-sizing.</p>
            </div>

            <div className="max-w-4xl mx-auto bg-white border border-gray-200 rounded-2xl p-8 shadow-sm grid md:grid-cols-12 gap-8 items-center">
              {/* Controls */}
              <div className="md:col-span-7 space-y-6">
                <div>
                  <label className="block text-xs font-bold text-gray-500 uppercase mb-2">
                    Cloud Infrastructure Provider
                  </label>
                  <div className="grid grid-cols-3 gap-3">
                    <button
                      onClick={() => setCloudProvider("aws")}
                      className={`py-2.5 text-xs font-bold rounded-xl border transition ${
                        cloudProvider === "aws" ? "bg-black text-white border-black" : "bg-gray-50 border-gray-200 text-gray-700"
                      }`}
                    >
                      Amazon AWS
                    </button>
                    <button
                      onClick={() => setCloudProvider("gcp")}
                      className={`py-2.5 text-xs font-bold rounded-xl border transition ${
                        cloudProvider === "gcp" ? "bg-black text-white border-black" : "bg-gray-50 border-gray-200 text-gray-700"
                      }`}
                    >
                      Google Cloud
                    </button>
                    <button
                      onClick={() => setCloudProvider("azure")}
                      className={`py-2.5 text-xs font-bold rounded-xl border transition ${
                        cloudProvider === "azure" ? "bg-black text-white border-black" : "bg-gray-50 border-gray-200 text-gray-700"
                      }`}
                    >
                      MS Azure
                    </button>
                  </div>
                </div>

                <div>
                  <div className="flex justify-between items-center mb-2">
                    <label className="text-xs font-bold text-gray-500 uppercase">Number of Active Worker Nodes</label>
                    <span className="text-lg font-extrabold text-blue-600">{nodeCount} Nodes</span>
                  </div>
                  <input
                    type="range"
                    min="5"
                    max="300"
                    step="5"
                    value={nodeCount}
                    onChange={(e) => setNodeCount(Number(e.target.value))}
                    className="w-full h-2 bg-gray-200 rounded-lg appearance-none cursor-pointer accent-black"
                  />
                  <div className="flex justify-between text-[11px] text-gray-400 mt-1">
                    <span>5 Nodes</span>
                    <span>150 Nodes</span>
                    <span>300+ Nodes</span>
                  </div>
                </div>
              </div>

              {/* Savings Output Box */}
              <div className="md:col-span-5 bg-gradient-to-br from-emerald-600 to-teal-700 text-white rounded-xl p-6 shadow-md text-center">
                <p className="text-xs uppercase font-bold tracking-wider text-emerald-100 mb-1">Estimated Monthly Savings</p>
                <p className="text-4xl font-black mb-4">${monthlySavings.toLocaleString()}</p>
                <div className="border-t border-emerald-500/50 pt-4 text-xs font-medium space-y-2">
                  <div className="flex justify-between">
                    <span className="text-emerald-100">Annual Return:</span>
                    <span className="font-bold">${annualSavings.toLocaleString()}</span>
                  </div>
                  <div className="flex justify-between">
                    <span className="text-emerald-100">Cost Reduction:</span>
                    <span className="font-bold">~38.0%</span>
                  </div>
                </div>
                <Link
                  href="/dashboard"
                  className="mt-6 block w-full py-2.5 bg-white text-emerald-800 font-bold text-xs rounded-lg hover:bg-emerald-50 transition shadow-sm"
                >
                  Start Saving Now →
                </Link>
              </div>
            </div>
          </div>
        </section>

        {/* 3-Layer Architecture Overview */}
        <section id="architecture" className="py-20 bg-white border-t border-black/10">
          <div className="max-w-7xl mx-auto px-6">
            <div className="text-center max-w-3xl mx-auto mb-16">
              <span className="text-xs font-bold text-purple-600 uppercase tracking-widest">Robust Architecture</span>
              <h2 className="text-3xl sm:text-4xl font-extrabold tracking-tight mt-2 mb-4">
                Designed for Zero Latency & Maximum Security
              </h2>
              <p className="text-gray-600 text-sm">
                Clean decoupling between cluster-bound DaemonSet agents and cloud-hosted decision solvers.
              </p>
            </div>

            <div className="grid md:grid-cols-3 gap-8">
              <div className="p-8 bg-[#F8F6F0] rounded-2xl border border-gray-200 shadow-sm relative">
                <span className="w-8 h-8 rounded-full bg-black text-white text-xs font-bold flex items-center justify-center mb-4">1</span>
                <h3 className="text-lg font-bold mb-2">In-Cluster Agent (Go)</h3>
                <p className="text-sm text-gray-600 leading-relaxed mb-4">
                  Deploys as a lightweight DaemonSet. Scrapes node metrics via client-go and streams encrypted telemetry over gRPC to the control plane every 15s.
                </p>
                <span className="text-[11px] font-mono text-gray-500 bg-white px-2 py-1 rounded border border-gray-200">
                  services/agent
                </span>
              </div>

              <div className="p-8 bg-[#F8F6F0] rounded-2xl border border-gray-200 shadow-sm relative">
                <span className="w-8 h-8 rounded-full bg-blue-600 text-white text-xs font-bold flex items-center justify-center mb-4">2</span>
                <h3 className="text-lg font-bold mb-2">AI Engine & Solvers</h3>
                <p className="text-sm text-gray-600 leading-relaxed mb-4">
                  FastAPI service hosting PuLP integer solvers and XGBoost inference models. Processes metrics from NATS JetStream and generates routing decisions.
                </p>
                <span className="text-[11px] font-mono text-gray-500 bg-white px-2 py-1 rounded border border-gray-200">
                  services/ai-engine
                </span>
              </div>

              <div className="p-8 bg-[#F8F6F0] rounded-2xl border border-gray-200 shadow-sm relative">
                <span className="w-8 h-8 rounded-full bg-emerald-600 text-white text-xs font-bold flex items-center justify-center mb-4">3</span>
                <h3 className="text-lg font-bold mb-2">Control Orchestrator</h3>
                <p className="text-sm text-gray-600 leading-relaxed mb-4">
                  Go-based REST & gRPC API server connected to PostgreSQL and Redis. Enforces human-in-the-loop approval rules and credit billing.
                </p>
                <span className="text-[11px] font-mono text-gray-500 bg-white px-2 py-1 rounded border border-gray-200">
                  services/orchestrator
                </span>
              </div>
            </div>
          </div>
        </section>

        {/* Pricing Section */}
        <section id="pricing" className="py-20 bg-[#F8F6F0] border-t border-black/10">
          <div className="max-w-7xl mx-auto px-6">
            <div className="text-center max-w-3xl mx-auto mb-16">
              <span className="text-xs font-bold text-blue-600 uppercase tracking-widest">Simple Credit Pricing</span>
              <h2 className="text-3xl sm:text-4xl font-extrabold tracking-tight mt-2 mb-4">
                Pay per Task Credit, Not Per Node
              </h2>
              <p className="text-gray-600 text-sm">
                Each automated routing decision costs 1 task credit. No hidden per-socket or per-core licensing fees.
              </p>
            </div>

            <div className="grid md:grid-cols-3 gap-8 max-w-5xl mx-auto">
              {/* Free Tier */}
              <div className="bg-white border border-gray-200 rounded-2xl p-8 shadow-sm flex flex-col justify-between">
                <div>
                  <h3 className="text-lg font-bold mb-1">Developer</h3>
                  <p className="text-xs text-gray-500 mb-6">For local development & Kind clusters</p>
                  <div className="mb-6">
                    <span className="text-4xl font-black">$0</span>
                    <span className="text-xs text-gray-500 ml-1">/ forever</span>
                  </div>
                  <ul className="space-y-3 text-xs text-gray-600 mb-8">
                    <li className="flex items-center gap-2"><span className="text-emerald-500 font-bold">✓</span> 100 Task Credits / month</li>
                    <li className="flex items-center gap-2"><span className="text-emerald-500 font-bold">✓</span> Up to 2 Clusters</li>
                    <li className="flex items-center gap-2"><span className="text-emerald-500 font-bold">✓</span> Manual & Auto Decision approvals</li>
                    <li className="flex items-center gap-2"><span className="text-emerald-500 font-bold">✓</span> Community Discord Support</li>
                  </ul>
                </div>
                <Link
                  href="/dashboard"
                  className="block w-full py-3 bg-gray-100 hover:bg-gray-200 text-gray-800 text-xs font-bold text-center rounded-xl transition"
                >
                  Start Free
                </Link>
              </div>

              {/* Pro Tier */}
              <div className="bg-white border-2 border-black rounded-2xl p-8 shadow-xl flex flex-col justify-between relative scale-105">
                <span className="absolute -top-3.5 left-1/2 -translate-x-1/2 px-3 py-1 bg-black text-white text-[10px] font-bold uppercase rounded-full tracking-wider">
                  Most Popular
                </span>
                <div>
                  <h3 className="text-lg font-bold mb-1">Production Pro</h3>
                  <p className="text-xs text-gray-500 mb-6">For mission-critical production workloads</p>
                  <div className="mb-6">
                    <span className="text-4xl font-black">$49</span>
                    <span className="text-xs text-gray-500 ml-1">/ cluster / mo</span>
                  </div>
                  <ul className="space-y-3 text-xs text-gray-600 mb-8">
                    <li className="flex items-center gap-2"><span className="text-emerald-500 font-bold">✓</span> 1,000 Task Credits included</li>
                    <li className="flex items-center gap-2"><span className="text-emerald-500 font-bold">✓</span> Unlimited Clusters</li>
                    <li className="flex items-center gap-2"><span className="text-emerald-500 font-bold">✓</span> Spot Cost Optimizer (PuLP)</li>
                    <li className="flex items-center gap-2"><span className="text-emerald-500 font-bold">✓</span> XGBoost Failure Prediction</li>
                    <li className="flex items-center gap-2"><span className="text-emerald-500 font-bold">✓</span> Priority Slack Support</li>
                  </ul>
                </div>
                <Link
                  href="/dashboard"
                  className="block w-full py-3 bg-black hover:bg-gray-800 text-white text-xs font-bold text-center rounded-xl transition shadow-md"
                >
                  Deploy Pro Cluster →
                </Link>
              </div>

              {/* Enterprise Tier */}
              <div className="bg-white border border-gray-200 rounded-2xl p-8 shadow-sm flex flex-col justify-between">
                <div>
                  <h3 className="text-lg font-bold mb-1">Enterprise</h3>
                  <p className="text-xs text-gray-500 mb-6">For large multi-cloud deployments</p>
                  <div className="mb-6">
                    <span className="text-4xl font-black">Custom</span>
                  </div>
                  <ul className="space-y-3 text-xs text-gray-600 mb-8">
                    <li className="flex items-center gap-2"><span className="text-emerald-500 font-bold">✓</span> Unlimited Task Credits</li>
                    <li className="flex items-center gap-2"><span className="text-emerald-500 font-bold">✓</span> On-Premise Air-Gapped Control Plane</li>
                    <li className="flex items-center gap-2"><span className="text-emerald-500 font-bold">✓</span> Custom Model Fine-Tuning</li>
                    <li className="flex items-center gap-2"><span className="text-emerald-500 font-bold">✓</span> 99.99% SLA & 24/7 Dedicated SRE</li>
                  </ul>
                </div>
                <a
                  href="mailto:enterprise@kuberoute.ai"
                  className="block w-full py-3 bg-gray-100 hover:bg-gray-200 text-gray-800 text-xs font-bold text-center rounded-xl transition"
                >
                  Contact Sales
                </a>
              </div>
            </div>
          </div>
        </section>

        {/* FAQ Section */}
        <section id="faq" className="py-20 bg-white border-t border-black/10">
          <div className="max-w-4xl mx-auto px-6">
            <div className="text-center mb-12">
              <span className="text-xs font-bold text-blue-600 uppercase tracking-widest">Got Questions?</span>
              <h2 className="text-3xl font-extrabold tracking-tight mt-2">Frequently Asked Questions</h2>
            </div>

            <div className="space-y-4">
              {[
                {
                  q: "How does Kuberoute AI integrate with existing ingress controllers (nginx, Traefik, Cilium)?",
                  a: "Kuberoute AI runs alongside your existing ingress controllers and CNI. It instructs the kube-scheduler and node affinity rules dynamically without requiring custom CRDs or replacing your network overlay."
                },
                {
                  q: "What are Task Credits?",
                  a: "A task credit represents 1 automated routing decision executed on your cluster (such as relocating a pod to a spot node or pre-scaling a deployment). Free tier includes 100 credits/month."
                },
                {
                  q: "What happens if the Cloud AI Control Plane loses network connection to the Agent?",
                  a: "The in-cluster Agent fallback handler automatically maintains existing pod routing rules. Kuberoute AI never disrupts running workloads during transient network disconnections."
                },
                {
                  q: "Does Kuberoute AI require cluster-admin privileges?",
                  a: "No. The agent requires RBAC permissions scoped only to read node metrics, watch pod statuses, and update pod nodeAffinity or deployment replicas."
                }
              ].map((item, idx) => (
                <div key={idx} className="border border-gray-200 rounded-xl overflow-hidden">
                  <button
                    onClick={() => setOpenFaq(openFaq === idx ? null : idx)}
                    className="w-full p-5 text-left font-bold text-sm flex justify-between items-center bg-[#F8F6F0]/50 hover:bg-[#F8F6F0] transition"
                  >
                    <span>{item.q}</span>
                    <span className="text-lg font-normal text-gray-400">{openFaq === idx ? "−" : "+"}</span>
                  </button>
                  {openFaq === idx && (
                    <div className="p-5 border-t border-gray-100 bg-white text-xs text-gray-600 leading-relaxed">
                      {item.a}
                    </div>
                  )}
                </div>
              ))}
            </div>
          </div>
        </section>

        {/* Bottom CTA Banner */}
        <section className="py-16 bg-black text-white text-center">
          <div className="max-w-4xl mx-auto px-6">
            <h2 className="text-3xl sm:text-4xl font-extrabold mb-4">Ready to optimize your Kubernetes clusters?</h2>
            <p className="text-gray-400 text-sm mb-8 max-w-xl mx-auto">
              Get started in under 3 minutes with our single-command Helm agent deployment.
            </p>
            <div className="flex flex-wrap justify-center gap-4">
              <Link
                href="/dashboard"
                className="px-6 py-3 bg-white text-black font-bold text-sm rounded-xl hover:bg-gray-100 transition shadow-lg"
              >
                Launch Dashboard →
              </Link>
              <a
                href={`${backendUrl}/auth/github`}
                className="px-6 py-3 bg-gray-800 text-white font-bold text-sm rounded-xl hover:bg-gray-700 transition border border-gray-700"
              >
                Sign In with GitHub
              </a>
            </div>
          </div>
        </section>
      </main>

      {/* Footer */}
      <footer className="border-t border-gray-200 bg-white py-12">
        <div className="max-w-7xl mx-auto px-6 flex flex-col sm:flex-row justify-between items-center gap-6">
          <div className="flex items-center gap-2 font-bold text-base">
            <div className="w-6 h-6 bg-black text-white rounded flex items-center justify-center text-xs font-bold">K</div>
            Kuberoute AI
          </div>

          <div className="flex gap-6 text-xs text-gray-500 font-medium">
            <a href="#visualizer" className="hover:text-black transition">3D Mesh</a>
            <a href="#features" className="hover:text-black transition">Features</a>
            <a href="#calculator" className="hover:text-black transition">ROI Calculator</a>
            <a href="#architecture" className="hover:text-black transition">Architecture</a>
            <a href="#pricing" className="hover:text-black transition">Pricing</a>
            <a href="#faq" className="hover:text-black transition">FAQ</a>
          </div>

          <div className="text-xs text-gray-400">
            © 2026 Kuberoute AI. All rights reserved.
          </div>
        </div>
      </footer>
    </div>
  );
}
