import type { Metadata } from "next";
import "./globals.css";

export const metadata: Metadata = {
  title: "Kuberoute AI — Autonomous Predictive Kubernetes Traffic & Cost Control Plane",
  description: "AI that predicts pod placement, eliminates Kubernetes over-provisioning, safely migrates workloads to spot instances, and prevents cluster outages before they occur.",
  keywords: [
    "Kubernetes AI",
    "K8s Traffic Routing",
    "Predictive Pod Placement",
    "Kubernetes Cost Optimization",
    "Spot Instance Manager",
    "Autonomous Infrastructure",
    "Cloud Native AI",
    "DevOps Automation"
  ],
  authors: [{ name: "Kuberoute AI Team" }],
  openGraph: {
    title: "Kuberoute AI — Autonomous Predictive Kubernetes Control Plane",
    description: "Predict traffic spikes, cut cloud compute costs by 40%, and eliminate Kubernetes outages automatically.",
    url: "https://kuberoute.ai",
    siteName: "Kuberoute AI",
    locale: "en_US",
    type: "website",
  },
  twitter: {
    card: "summary_large_image",
    title: "Kuberoute AI — Predictive Control Plane for Kubernetes",
    description: "Autonomous pod routing, spot cost optimization, and proactive outage prevention.",
  },
  robots: {
    index: true,
    follow: true,
  },
};

export default function RootLayout({ children }: { children: React.ReactNode }) {
  return (
    <html lang="en" className="scroll-smooth">
      <head>
        <link rel="canonical" href="https://kuberoute.ai" />
        <script
          dangerouslySetInnerHTML={{
            __html: `
              if (typeof window !== 'undefined' && 'serviceWorker' in navigator) {
                navigator.serviceWorker.getRegistrations().then(function(registrations) {
                  for (var registration of registrations) {
                    registration.unregister();
                  }
                });
              }
            `,
          }}
        />
        <script
          type="application/ld+json"
          dangerouslySetInnerHTML={{
            __html: JSON.stringify({
              "@context": "https://schema.org",
              "@type": "SoftwareApplication",
              "name": "Kuberoute AI",
              "applicationCategory": "DevOpsApplication",
              "operatingSystem": "Kubernetes",
              "offers": {
                "@type": "Offer",
                "price": "0",
                "priceCurrency": "USD"
              },
              "description": "Autonomous AI control plane for Kubernetes traffic routing and cost optimization."
            }),
          }}
        />
      </head>
      <body className="bg-[#F8F6F0] text-[#1A1A1A] font-sans antialiased selection:bg-blue-600 selection:text-white">
        {children}
      </body>
    </html>
  );
}
