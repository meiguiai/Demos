//
//  ContentView.swift
//  FBGroupsSwiftUI
//
//  Created by ylw on 2019/6/17.
//  Copyright © 2019 ylw. All rights reserved.
//

import SwiftUI

struct Post {
    let id: Int
    let username, text, imagename, useIcon: String
}

struct ContentView : View {
    
    let posts = [
        Post.init(id: 0, username: "Enter", text: "Good old bill up to his usual ways and dirty tricks", imagename: "55", useIcon: "5"),
        Post.init(id: 1, username: "Enter", text: "Good old bill up to his usual ways and dirty tricks", imagename: "22", useIcon: "2"),
        Post.init(id: 3, username: "Enter", text: "Good old bill up to his usual ways and dirty tricks", imagename: "11", useIcon: "1"),
        Post.init(id: 2, username: "Enter", text: "Good old bill up to his usual ways and dirty tricks", imagename: "44", useIcon: "4")
    ]
    
    var body: some View {
        NavigationView {
            List {
                VStack(alignment:.leading) {
                    Text(verbatim:"Trending")
                    ScrollView {
                        VStack (alignment: .leading) {
                            HStack {
                                ForEach(posts.identified(by: \.id)) { post in
                                    NavigationButton(destination: GroupDetailView()) {
                                        GroupView(post: post)
                                    }
                                }
                            }
                        }
                        }.frame(height:200)
                }
                ForEach(posts.identified(by: \.id)) { post in
                    PostView(post: post)
                }
            }.navigationBarTitle(Text("Groups"))
        }
    }
}

struct GroupDetailView: View {
    var body: some View {
        Text("Group Detail View ")
    }
}

struct GroupView: View {
    
    let post: Post
    var body: some View {
        VStack {
            Image(post.useIcon)
                .resizable()
                .renderingMode(.original) //解决默认蓝色背景问题
                .cornerRadius(12)
                .frame(width:80,height: 80)
            Text("Hstack of colorado")
                .color(.primary)
                .lineLimit(nil)
                .padding(.leading, 0)
            }.frame(width:100,height: 170).padding(.leading,0)
    }
}

struct PostView: View {
    let post: Post
    var body: some View {
        VStack(alignment: .leading, spacing: 12) {
            HStack {
                Image("33")
                    .resizable()
                    .clipShape(Circle())
                    .frame(width:60,height: 60)
                    .clipped()
                VStack(alignment: .leading,spacing: 4) {
                    Text(post.username).font(.headline)
                    Text("Enter-ylw").font(.subheadline)
                }.padding(.leading,8)
            }.padding(.leading,16).padding(.top,16)
            Text(post.text).lineLimit(nil).padding(.leading,16).padding(.trailing,16)
            Image(post.imagename)
                .scaledToFill()
                .frame(height:360)
                .clipped()
        }.padding(.leading,-20).padding(.bottom,-8)
    }
}

#if DEBUG
struct ContentView_Previews : PreviewProvider {
    static var previews: some View {
        ContentView()
    }
}
#endif
