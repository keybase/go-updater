//
//  Prompt.h
//  Updater
//
//  Created by Gabriel on 4/13/16.
//  Copyright © 2016 Gabriel Handford. All rights reserved.
//

#import <Foundation/Foundation.h>

#import <AppKit/AppKit.h>

@interface Prompt : NSObject

+ (void)showPromptWithInputString:(NSString *)inputString presenter:(NSModalResponse (^)(NSAlert *alert))presenter completion:(void (^)(NSError *error, NSData *output))completion;
+ (void)showUpdatePrompt:(NSDictionary *)input presenter:(NSModalResponse (^)(NSAlert *alert))presenter completion:(void (^)(NSError *error, NSData *output))completion;

@end
