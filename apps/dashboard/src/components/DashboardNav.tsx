"use client";

import Link from "next/link";
import { usePathname } from "next/navigation";

const links = [
  { href: "/dashboard", label: "Overview" },
  { href: "/dashboard/decisions", label: "Decisions" },
  { href: "/dashboard/costs", label: "Costs" },
  { href: "/dashboard/settings", label: "Settings" },
];

export default function DashboardNav() {
  const path = usePathname();
  return (
    <header className="border-b border-black/10 bg-white sticky top-0 z-50">
      <div className="max-w-7xl mx-auto px-6 h-16 flex items-center justify-between">
        <div className="flex items-center gap-6">
          <Link href="/" className="flex items-center gap-2 font-bold text-lg">
            <div className="w-7 h-7 bg-black text-white rounded flex items-center justify-center text-sm font-bold">K</div>
            Kuberoute AI
          </Link>
          <nav className="flex items-center gap-1 text-sm font-medium text-gray-600">
            {links.map((l) => (
              <Link
                key={l.href}
                href={l.href}
                className={`px-3 py-1.5 rounded-lg transition ${
                  path === l.href || (l.href !== "/dashboard" && path.startsWith(l.href))
                    ? "bg-gray-100 text-black font-semibold"
                    : "hover:bg-gray-100"
                }`}
              >
                {l.label}
              </Link>
            ))}
          </nav>
        </div>
      </div>
    </header>
  );
}
