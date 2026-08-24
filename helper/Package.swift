// swift-tools-version: 6.0

import PackageDescription

let package = Package(
    name: "ocrogram-helper",
    platforms: [
        .macOS(.v14),
    ],
    products: [
        .executable(name: "ocrogram-helper", targets: ["ocrogram-helper"]),
    ],
    targets: [
        .executableTarget(name: "ocrogram-helper"),
    ]
)
