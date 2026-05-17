// swift-tools-version: 5.9
import PackageDescription

let package = Package(
    name: "AsunConformanceSwift",
    platforms: [
        .macOS(.v13)
    ],
    dependencies: [
        .package(path: "../../../asun-swift")
    ],
    targets: [
        .executableTarget(
            name: "asun-conformance-swift",
            dependencies: [
                .product(name: "AsunSwift", package: "asun-swift")
            ],
            path: "Sources/asun-conformance-swift"
        )
    ]
)
