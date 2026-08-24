import Foundation
import Vision

func recognizeText(in imageURL: URL) throws -> String {
    let request = VNRecognizeTextRequest()
    request.recognitionLevel = .accurate
    request.usesLanguageCorrection = true
    request.automaticallyDetectsLanguage = true

    let handler = VNImageRequestHandler(url: imageURL)
    try handler.perform([request])

    let lines = (request.results ?? []).compactMap { observation in
        observation.topCandidates(1).first?.string
    }
    return lines.joined(separator: "\n")
}
