//import SwiftUI
//import Foundation
//
//struct GlassCard<Content: View>: View {
//    var padding: CGFloat = 16
//    @ViewBuilder let content: Content
//
//    var body: some View {
//        content
//            .padding(padding)
//            .background(
//                RoundedRectangle(cornerRadius: 22, style: .continuous)
//                    .fill(Color.white.opacity(0.07))
//                    .overlay(
//                        RoundedRectangle(cornerRadius: 22, style: .continuous)
//                            .stroke(Color.white.opacity(0.10), lineWidth: 1)
//                    )
//                    .shadow(color: .black.opacity(0.18), radius: 18, x: 0, y: 10)
//            )
//    }
//}
//
//struct SectionCard<Content: View>: View {
//    let title: String
//    let icon: String
//    @ViewBuilder let content: Content
//
//    var body: some View {
//        GlassCard {
//            VStack(alignment: .leading, spacing: 14) {
//                HStack(spacing: 10) {
//                    Image(systemName: icon)
//                        .font(.system(size: 15, weight: .semibold))
//                        .foregroundStyle(Color(red: 0.0, green: 0.65, blue: 1.0))
//                        .frame(width: 28, height: 28)
//                        .background(Color.white.opacity(0.06))
//                        .clipShape(RoundedRectangle(cornerRadius: 10, style: .continuous))
//
//                    Text(title)
//                        .font(.system(size: 17, weight: .bold, design: .rounded))
//                        .foregroundStyle(.white)
//
//                    Spacer(minLength: 0)
//                }
//
//                content
//            }
//        }
//    }
//}
//
//struct EmptyStateView: View {
//    let icon: String
//    let title: String
//    let subtitle: String
//
//    var body: some View {
//        VStack(spacing: 8) {
//            Image(systemName: icon)
//                .font(.system(size: 20, weight: .semibold))
//                .foregroundStyle(.white.opacity(0.7))
//                .frame(width: 44, height: 44)
//                .background(Color.white.opacity(0.05))
//                .clipShape(RoundedRectangle(cornerRadius: 14, style: .continuous))
//
//            Text(title)
//                .font(.system(size: 15, weight: .semibold))
//                .foregroundStyle(.white)
//
//            Text(subtitle)
//                .font(.system(size: 13))
//                .foregroundStyle(.white.opacity(0.5))
//                .multilineTextAlignment(.center)
//        }
//        .frame(maxWidth: .infinity)
//        .padding(.vertical, 12)
//    }
//}
//
