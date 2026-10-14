"use client";

import { useEffect, useRef, useState } from "react";
import * as THREE from "three";

export default function Cluster3DCanvas() {
  const mountRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    const container = mountRef.current;
    if (!container) return;

    // Scene setup
    const scene = new THREE.Scene();
    scene.fog = new THREE.FogExp2(0x0a0a10, 0.025);

    // Camera setup - Centered directly on origin (0, 0, 0)
    const camera = new THREE.PerspectiveCamera(
      55,
      container.clientWidth / container.clientHeight,
      0.1,
      1000
    );
    camera.position.set(0, 0, 11);
    camera.lookAt(0, 0, 0);

    // Renderer setup
    const renderer = new THREE.WebGLRenderer({ antialias: true, alpha: true });
    renderer.setSize(container.clientWidth, container.clientHeight);
    renderer.setPixelRatio(Math.min(window.devicePixelRatio, 2));
    container.appendChild(renderer.domElement);

    // Ambient & Point Lighting
    const ambientLight = new THREE.AmbientLight(0xffffff, 0.8);
    scene.add(ambientLight);

    const mainLight = new THREE.PointLight(0x3b82f6, 4, 60);
    mainLight.position.set(0, 5, 12);
    scene.add(mainLight);

    const emeraldLight = new THREE.PointLight(0x10b981, 3, 50);
    emeraldLight.position.set(-10, -5, 5);
    scene.add(emeraldLight);

    const purpleLight = new THREE.PointLight(0x8b5cf6, 3, 50);
    purpleLight.position.set(10, 5, -5);
    scene.add(purpleLight);

    // Centered Cluster Group
    const clusterGroup = new THREE.Group();
    clusterGroup.position.set(0, 0, 0);
    scene.add(clusterGroup);

    // Larger, Centered Node Mesh Layout
    const nodePositions = [
      { id: "node-1", x: -4.5, y: 1.8, z: 0, color: 0x3b82f6, name: "Node A (Spot e2-standard-8)" },
      { id: "node-2", x: 4.5, y: 1.8, z: 0, color: 0x8b5cf6, name: "Node B (Spot m5.xlarge)" },
      { id: "node-3", x: -2.8, y: -2.5, z: 1.5, color: 0x10b981, name: "Node C (On-Demand c2-standard-8)" },
      { id: "node-4", x: 2.8, y: -2.5, z: -1.5, color: 0x06b6d4, name: "Node D (Spot c5.2xlarge)" },
      { id: "node-center", x: 0, y: 0, z: 0, color: 0xec4899, name: "Control Plane Master Node" },
    ];

    const nodeMeshes: THREE.Mesh[] = [];

    nodePositions.forEach((pos) => {
      // Bigger Node Cube (size 2.0 vs 1.4)
      const size = pos.id === "node-center" ? 2.4 : 2.0;
      const geometry = new THREE.BoxGeometry(size, size, size);
      const material = new THREE.MeshStandardMaterial({
        color: pos.color,
        roughness: 0.15,
        metalness: 0.85,
      });

      const mesh = new THREE.Mesh(geometry, material);
      mesh.position.set(pos.x, pos.y, pos.z);
      mesh.userData = pos;
      clusterGroup.add(mesh);
      nodeMeshes.push(mesh);

      // Outer Wireframe Cage (size + 0.3)
      const wireSize = size + 0.35;
      const wireGeo = new THREE.BoxGeometry(wireSize, wireSize, wireSize);
      const wireMat = new THREE.MeshBasicMaterial({
        color: pos.color,
        wireframe: true,
        transparent: true,
        opacity: 0.45,
      });
      const wireMesh = new THREE.Mesh(wireGeo, wireMat);
      mesh.add(wireMesh);

      // Orbiting Pod Spheres
      const podCount = pos.id === "node-center" ? 6 : 4;
      for (let i = 0; i < podCount; i++) {
        const podGeo = new THREE.SphereGeometry(0.28, 16, 16);
        const podMat = new THREE.MeshStandardMaterial({
          color: 0xffffff,
          emissive: pos.color,
          emissiveIntensity: 0.8,
        });
        const podMesh = new THREE.Mesh(podGeo, podMat);
        const angle = (i / podCount) * Math.PI * 2;
        const radius = size * 1.2;
        podMesh.position.set(Math.cos(angle) * radius, Math.sin(angle) * radius, (i - podCount / 2) * 0.4);
        mesh.add(podMesh);
      }
    });

    // Connecting Energy Lines from Center to Surrounding Nodes
    const centerPos = new THREE.Vector3(0, 0, 0);
    nodePositions.forEach((pos) => {
      if (pos.id === "node-center") return;
      const targetPos = new THREE.Vector3(pos.x, pos.y, pos.z);

      const lineGeo = new THREE.BufferGeometry().setFromPoints([centerPos, targetPos]);
      const lineMat = new THREE.LineBasicMaterial({
        color: pos.color,
        transparent: true,
        opacity: 0.6,
        linewidth: 2,
      });
      const line = new THREE.Line(lineGeo, lineMat);
      clusterGroup.add(line);
    });

    // Ambient Background Particles
    const particleCount = 250;
    const particleGeo = new THREE.BufferGeometry();
    const particlePositions = new Float32Array(particleCount * 3);

    for (let i = 0; i < particleCount * 3; i += 3) {
      particlePositions[i] = (Math.random() - 0.5) * 35;
      particlePositions[i + 1] = (Math.random() - 0.5) * 25;
      particlePositions[i + 2] = (Math.random() - 0.5) * 25;
    }

    particleGeo.setAttribute("position", new THREE.BufferAttribute(particlePositions, 3));
    const particleMat = new THREE.PointsMaterial({
      color: 0xa5b4fc,
      size: 0.1,
      transparent: true,
      opacity: 0.65,
    });

    const particles = new THREE.Points(particleGeo, particleMat);
    scene.add(particles);

    // Mouse Interaction
    let mouseX = 0;
    let mouseY = 0;

    const handleMouseMove = (event: MouseEvent) => {
      const rect = container.getBoundingClientRect();
      mouseX = ((event.clientX - rect.left) / container.clientWidth) * 2 - 1;
      mouseY = -((event.clientY - rect.top) / container.clientHeight) * 2 + 1;
    };

    container.addEventListener("mousemove", handleMouseMove);

    // Animation Loop
    let animationFrameId: number;

    const animate = () => {
      animationFrameId = requestAnimationFrame(animate);

      // Smooth Gentle Rotation Centered at Origin
      clusterGroup.rotation.y += 0.004;
      clusterGroup.rotation.x = mouseY * 0.15;
      clusterGroup.rotation.y += mouseX * 0.005;

      // Rotate individual node cubes
      nodeMeshes.forEach((mesh) => {
        mesh.rotation.x += 0.008;
        mesh.rotation.y += 0.008;
      });

      // Animate background particles
      const positions = particles.geometry.attributes.position.array as Float32Array;
      for (let i = 1; i < particleCount * 3; i += 3) {
        positions[i] -= 0.02;
        if (positions[i] < -12) positions[i] = 12;
      }
      particles.geometry.attributes.position.needsUpdate = true;

      renderer.render(scene, camera);
    };

    animate();

    const handleResize = () => {
      if (!container) return;
      camera.aspect = container.clientWidth / container.clientHeight;
      camera.updateProjectionMatrix();
      renderer.setSize(container.clientWidth, container.clientHeight);
    };

    window.addEventListener("resize", handleResize);

    return () => {
      cancelAnimationFrame(animationFrameId);
      container.removeEventListener("mousemove", handleMouseMove);
      window.removeEventListener("resize", handleResize);
      if (container.contains(renderer.domElement)) {
        container.removeChild(renderer.domElement);
      }
    };
  }, []);

  return (
    <div className="relative w-full h-[520px] bg-[#0A0A10] rounded-2xl overflow-hidden border border-gray-800 shadow-2xl flex items-center justify-center">
      {/* 3D WebGL Canvas Container */}
      <div ref={mountRef} className="w-full h-full cursor-grab active:cursor-grabbing" />

      {/* Overlay Status Pill */}
      <div className="absolute top-4 left-4 pointer-events-none">
        <div className="flex items-center gap-2 px-3 py-1.5 rounded-full bg-blue-950/80 border border-blue-800/80 text-blue-300 text-xs font-mono shadow-md backdrop-blur-md">
          <span className="w-2 h-2 rounded-full bg-emerald-400 animate-ping" />
          <span>WebGL 3D Topology Mesh (Centered)</span>
        </div>
      </div>

      <div className="absolute bottom-4 right-4 pointer-events-none">
        <span className="text-[11px] font-mono text-gray-400 bg-black/70 px-3 py-1.5 rounded-lg border border-gray-800 backdrop-blur-md">
          Drag cursor to rotate • 5 Active Cluster Nodes
        </span>
      </div>
    </div>
  );
}
