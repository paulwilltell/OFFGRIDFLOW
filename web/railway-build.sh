#!/bin/bash
# Railway build script for offgridflow-web service

echo "Building OffGridFlow Web (Next.js)..."

cd web || exit 1

echo "Installing dependencies..."
npm ci

echo "Building Next.js application..."
npm run build

echo "Build complete!"
