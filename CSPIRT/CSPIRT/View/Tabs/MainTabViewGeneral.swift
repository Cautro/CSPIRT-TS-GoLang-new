import SwiftUI

struct MainTabView: View {
   @State private var selectedTab = 0
   let classId: Int

   var body: some View {
       TabView(selection: $selectedTab) {
           if classId != 0 {
               MyClassView()
                   .tabItem {
                       Image(systemName: "graduationcap")
                   }
                   .tag(1)

               ParallelsView()
                   .tabItem {
                       Image(systemName: "chart.bar")
                   }
                   .tag(2)

               MainPageView()
                   .tabItem {
                       Image(systemName: "house")
                   }
                   .tag(0)

               ScheduleView()
                   .tabItem {
                       Image(systemName: "calendar")
                   }
                   .tag(3)

               ProfileView()
                   .tabItem {
                       Image(systemName: "person.crop.circle")
                   }
                   .tag(4)
           } else {
               ParallelsView()
                   .tabItem {
                       Image(systemName: "chart.bar")
                   }
                   .tag(2)

               MainPageView()
                   .tabItem {
                       Image(systemName: "house")
                   }
                   .tag(0)
               ProfileView()
                   .tabItem {
                       Image(systemName: "person.crop.circle")
                   }
                   .tag(4)
           }
       }
       .tint(Color(red: 0.2, green: 0.65, blue: 1.0))
   }
}
