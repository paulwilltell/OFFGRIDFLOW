'use client';

import React, { useRef, useEffect, useState, memo, useCallback } from 'react';
import { DataSource } from '@/stores/carbonStore';

interface DataGlobeProps {
  nodes: DataSource[];
  onNodeClick?: (node: DataSource) => void;
}

interface Point3D {
  x: number;
  y: number;
  z: number;
}

interface GlobeNode extends DataSource {
  point: Point3D;
  screenX: number;
  screenY: number;
  scale: number;
  visible: boolean;
}

export const DataGlobe = memo(function DataGlobe({ nodes, onNodeClick }: DataGlobeProps) {
  const canvasRef = useRef<HTMLCanvasElement>(null);
  const containerRef = useRef<HTMLDivElement>(null);
  const animationRef = useRef<number>();
  const [hoveredNode, setHoveredNode] = useState<DataSource | null>(null);
  const [rotation, setRotation] = useState({ x: 0.3, y: 0 });
  const [isDragging, setIsDragging] = useState(false);
  const [lastMouse, setLastMouse] = useState({ x: 0, y: 0 });
  const [dimensions, setDimensions] = useState({ width: 0, height: 0 });
  const globeNodesRef = useRef<GlobeNode[]>([]);

  // Convert lat/lng to 3D point
  const latLngToPoint = useCallback((lat: number, lng: number, radius: number): Point3D => {
    const latRad = (lat * Math.PI) / 180;
    const lngRad = (lng * Math.PI) / 180;
    return {
      x: radius * Math.cos(latRad) * Math.cos(lngRad),
      y: radius * Math.sin(latRad),
      z: radius * Math.cos(latRad) * Math.sin(lngRad),
    };
  }, []);

  // Rotate point around axes
  const rotatePoint = useCallback((point: Point3D, rotX: number, rotY: number): Point3D => {
    // Rotate around Y axis
    let x = point.x * Math.cos(rotY) - point.z * Math.sin(rotY);
    let z = point.x * Math.sin(rotY) + point.z * Math.cos(rotY);
    let y = point.y;

    // Rotate around X axis
    const newY = y * Math.cos(rotX) - z * Math.sin(rotX);
    const newZ = y * Math.sin(rotX) + z * Math.cos(rotX);

    return { x, y: newY, z: newZ };
  }, []);

  // Project 3D to 2D
  const projectPoint = useCallback((point: Point3D, centerX: number, centerY: number, perspective: number): { x: number; y: number; scale: number } => {
    const scale = perspective / (perspective + point.z);
    return {
      x: centerX + point.x * scale,
      y: centerY - point.y * scale,
      scale,
    };
  }, []);

  // Update dimensions on resize
  useEffect(() => {
    const updateDimensions = () => {
      if (containerRef.current) {
        const rect = containerRef.current.getBoundingClientRect();
        setDimensions({ width: rect.width, height: rect.height });
      }
    };

    updateDimensions();
    window.addEventListener('resize', updateDimensions);
    return () => window.removeEventListener('resize', updateDimensions);
  }, []);

  // Initialize globe nodes from data sources
  useEffect(() => {
    const radius = Math.min(dimensions.width, dimensions.height) * 0.35;
    
    globeNodesRef.current = nodes.map((node) => {
      const coords = node.coordinates || { 
        lat: Math.random() * 140 - 70, 
        lng: Math.random() * 360 - 180 
      };
      const point = latLngToPoint(coords.lat, coords.lng, radius);
      
      return {
        ...node,
        point,
        screenX: 0,
        screenY: 0,
        scale: 1,
        visible: true,
      };
    });
  }, [nodes, dimensions, latLngToPoint]);

  // Draw the globe
  const draw = useCallback((ctx: CanvasRenderingContext2D, width: number, height: number) => {
    const centerX = width / 2;
    const centerY = height / 2;
    const radius = Math.min(width, height) * 0.35;
    const perspective = 400;

    // Clear canvas
    ctx.clearRect(0, 0, width, height);

    // Draw globe background
    const gradient = ctx.createRadialGradient(centerX, centerY, 0, centerX, centerY, radius);
    gradient.addColorStop(0, 'rgba(34, 197, 94, 0.1)');
    gradient.addColorStop(0.5, 'rgba(34, 197, 94, 0.05)');
    gradient.addColorStop(1, 'rgba(34, 197, 94, 0)');
    
    ctx.beginPath();
    ctx.arc(centerX, centerY, radius, 0, Math.PI * 2);
    ctx.fillStyle = gradient;
    ctx.fill();

    // Draw globe outline
    ctx.beginPath();
    ctx.arc(centerX, centerY, radius, 0, Math.PI * 2);
    ctx.strokeStyle = 'rgba(34, 197, 94, 0.3)';
    ctx.lineWidth = 2;
    ctx.stroke();

    // Draw latitude lines
    for (let lat = -60; lat <= 60; lat += 30) {
      ctx.beginPath();
      const latRadius = radius * Math.cos((lat * Math.PI) / 180);
      const latY = radius * Math.sin((lat * Math.PI) / 180);
      
      for (let lng = 0; lng <= 360; lng += 5) {
        const point = latLngToPoint(lat, lng, radius);
        const rotated = rotatePoint(point, rotation.x, rotation.y);
        
        if (rotated.z > 0) {
          const projected = projectPoint(rotated, centerX, centerY, perspective);
          if (lng === 0) {
            ctx.moveTo(projected.x, projected.y);
          } else {
            ctx.lineTo(projected.x, projected.y);
          }
        }
      }
      ctx.strokeStyle = 'rgba(34, 197, 94, 0.15)';
      ctx.lineWidth = 1;
      ctx.stroke();
    }

    // Draw longitude lines
    for (let lng = 0; lng < 360; lng += 30) {
      ctx.beginPath();
      for (let lat = -90; lat <= 90; lat += 5) {
        const point = latLngToPoint(lat, lng, radius);
        const rotated = rotatePoint(point, rotation.x, rotation.y);
        
        if (rotated.z > 0) {
          const projected = projectPoint(rotated, centerX, centerY, perspective);
          if (lat === -90) {
            ctx.moveTo(projected.x, projected.y);
          } else {
            ctx.lineTo(projected.x, projected.y);
          }
        }
      }
      ctx.strokeStyle = 'rgba(34, 197, 94, 0.15)';
      ctx.lineWidth = 1;
      ctx.stroke();
    }

    // Update and draw nodes
    globeNodesRef.current.forEach((node) => {
      const rotated = rotatePoint(node.point, rotation.x, rotation.y);
      const projected = projectPoint(rotated, centerX, centerY, perspective);
      
      node.screenX = projected.x;
      node.screenY = projected.y;
      node.scale = projected.scale;
      node.visible = rotated.z > 0;

      if (node.visible) {
        // Draw node
        const nodeRadius = 8 * projected.scale;
        const isHovered = hoveredNode?.id === node.id;
        
        // Glow effect
        const glowGradient = ctx.createRadialGradient(
          projected.x, projected.y, 0,
          projected.x, projected.y, nodeRadius * 3
        );
        
        const color = node.status === 'active' ? '34, 197, 94' : 
                     node.status === 'error' ? '239, 68, 68' : 
                     '234, 179, 8';
        
        glowGradient.addColorStop(0, `rgba(${color}, ${isHovered ? 0.5 : 0.3})`);
        glowGradient.addColorStop(1, `rgba(${color}, 0)`);
        
        ctx.beginPath();
        ctx.arc(projected.x, projected.y, nodeRadius * 3, 0, Math.PI * 2);
        ctx.fillStyle = glowGradient;
        ctx.fill();

        // Node circle
        ctx.beginPath();
        ctx.arc(projected.x, projected.y, nodeRadius, 0, Math.PI * 2);
        ctx.fillStyle = `rgba(${color}, ${isHovered ? 1 : 0.8})`;
        ctx.fill();
        ctx.strokeStyle = '#fff';
        ctx.lineWidth = isHovered ? 2 : 1;
        ctx.stroke();

        // Pulse animation for active nodes
        if (node.status === 'active') {
          const pulseRadius = nodeRadius * (1 + 0.5 * Math.sin(Date.now() / 500));
          ctx.beginPath();
          ctx.arc(projected.x, projected.y, pulseRadius, 0, Math.PI * 2);
          ctx.strokeStyle = `rgba(${color}, 0.5)`;
          ctx.lineWidth = 1;
          ctx.stroke();
        }
      }
    });

    // Draw connections between nodes
    ctx.beginPath();
    globeNodesRef.current.forEach((node, i) => {
      if (!node.visible) return;
      
      globeNodesRef.current.slice(i + 1).forEach((other) => {
        if (!other.visible) return;
        
        const dist = Math.sqrt(
          Math.pow(node.screenX - other.screenX, 2) +
          Math.pow(node.screenY - other.screenY, 2)
        );
        
        if (dist < 150) {
          ctx.moveTo(node.screenX, node.screenY);
          ctx.lineTo(other.screenX, other.screenY);
        }
      });
    });
    ctx.strokeStyle = 'rgba(34, 197, 94, 0.2)';
    ctx.lineWidth = 1;
    ctx.stroke();
  }, [rotation, hoveredNode, latLngToPoint, rotatePoint, projectPoint]);

  // Animation loop
  useEffect(() => {
    const canvas = canvasRef.current;
    if (!canvas || dimensions.width === 0) return;

    const ctx = canvas.getContext('2d');
    if (!ctx) return;

    const animate = () => {
      if (!isDragging) {
        setRotation((prev) => ({
          ...prev,
          y: prev.y + 0.002,
        }));
      }
      draw(ctx, dimensions.width, dimensions.height);
      animationRef.current = requestAnimationFrame(animate);
    };

    animate();

    return () => {
      if (animationRef.current) {
        cancelAnimationFrame(animationRef.current);
      }
    };
  }, [dimensions, isDragging, draw]);

  // Mouse handlers
  const handleMouseDown = (e: React.MouseEvent) => {
    setIsDragging(true);
    setLastMouse({ x: e.clientX, y: e.clientY });
  };

  const handleMouseMove = (e: React.MouseEvent) => {
    const rect = containerRef.current?.getBoundingClientRect();
    if (!rect) return;

    const mouseX = e.clientX - rect.left;
    const mouseY = e.clientY - rect.top;

    // Check for hovered node
    const hovered = globeNodesRef.current.find((node) => {
      if (!node.visible) return false;
      const dist = Math.sqrt(
        Math.pow(mouseX - node.screenX, 2) +
        Math.pow(mouseY - node.screenY, 2)
      );
      return dist < 15;
    });
    setHoveredNode(hovered || null);

    // Handle dragging
    if (isDragging) {
      const deltaX = e.clientX - lastMouse.x;
      const deltaY = e.clientY - lastMouse.y;
      
      setRotation((prev) => ({
        x: Math.max(-1.2, Math.min(1.2, prev.x + deltaY * 0.005)),
        y: prev.y + deltaX * 0.005,
      }));
      
      setLastMouse({ x: e.clientX, y: e.clientY });
    }
  };

  const handleMouseUp = () => {
    setIsDragging(false);
  };

  const handleClick = () => {
    if (hoveredNode && onNodeClick) {
      onNodeClick(hoveredNode);
    }
  };

  return (
    <div 
      ref={containerRef}
      className="relative bg-gray-800/50 rounded-xl border border-gray-700/50 overflow-hidden"
      style={{ height: 320 }}
    >
      {/* Header */}
      <div className="absolute top-4 left-4 z-10">
        <h3 className="text-sm font-semibold text-gray-400 uppercase tracking-wider">
          Data Sources
        </h3>
        <p className="text-xs text-gray-500 mt-1">
          {nodes.filter(n => n.status === 'active').length} of {nodes.length} active
        </p>
      </div>

      {/* Canvas */}
      <canvas
        ref={canvasRef}
        width={dimensions.width}
        height={dimensions.height}
        className="cursor-grab active:cursor-grabbing"
        onMouseDown={handleMouseDown}
        onMouseMove={handleMouseMove}
        onMouseUp={handleMouseUp}
        onMouseLeave={handleMouseUp}
        onClick={handleClick}
      />

      {/* Tooltip */}
      {hoveredNode && (
        <div
          className="absolute z-20 pointer-events-none"
          style={{
            left: hoveredNode.screenX || 0,
            top: (hoveredNode.screenY || 0) - 60,
            transform: 'translateX(-50%)',
          }}
        >
          <div className="bg-gray-900 border border-gray-700 rounded-lg p-3 shadow-xl min-w-[160px]">
            <div className="text-sm font-medium text-white mb-1">{hoveredNode.name}</div>
            <div className="flex items-center gap-2 text-xs">
              <span className={`px-1.5 py-0.5 rounded ${
                hoveredNode.status === 'active' ? 'bg-green-500/20 text-green-400' :
                hoveredNode.status === 'error' ? 'bg-red-500/20 text-red-400' :
                'bg-yellow-500/20 text-yellow-400'
              }`}>
                {hoveredNode.status}
              </span>
              <span className="text-gray-400">{hoveredNode.type}</span>
            </div>
            <div className="text-xs text-gray-500 mt-2">
              Last sync: {new Date(hoveredNode.lastSync).toLocaleTimeString()}
            </div>
          </div>
        </div>
      )}

      {/* Legend */}
      <div className="absolute bottom-4 right-4 flex items-center gap-4">
        <div className="flex items-center gap-1.5">
          <span className="w-2 h-2 rounded-full bg-green-500" />
          <span className="text-xs text-gray-400">Active</span>
        </div>
        <div className="flex items-center gap-1.5">
          <span className="w-2 h-2 rounded-full bg-yellow-500" />
          <span className="text-xs text-gray-400">Inactive</span>
        </div>
        <div className="flex items-center gap-1.5">
          <span className="w-2 h-2 rounded-full bg-red-500" />
          <span className="text-xs text-gray-400">Error</span>
        </div>
      </div>
    </div>
  );
});

export default DataGlobe;
