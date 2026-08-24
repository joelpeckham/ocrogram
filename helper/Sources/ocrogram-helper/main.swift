import Foundation

let usage = "usage: ocrogram-helper <image-path>"

guard CommandLine.arguments.count == 2 else {
    FileHandle.standardError.write(Data((usage + "\n").utf8))
    exit(2)
}

let path = CommandLine.arguments[1]
if path == "-h" || path == "--help" {
    print(usage)
    exit(0)
}

let url = URL(fileURLWithPath: path)
guard FileManager.default.fileExists(atPath: url.path) else {
    FileHandle.standardError.write(Data("ocrogram-helper: no such file: \(path)\n".utf8))
    exit(2)
}

let text: String
do {
    text = try recognizeText(in: url)
} catch {
    FileHandle.standardError.write(Data("ocrogram-helper: \(error.localizedDescription)\n".utf8))
    exit(2)
}

if text.isEmpty {
    exit(1)
}

copyToClipboard(text)
print(text)
exit(0)
