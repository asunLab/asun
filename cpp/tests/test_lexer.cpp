#include <gtest/gtest.h>
#include <ason/lexer.hpp>

using namespace ason;

TEST(LexerTest, EmptyInput) {
    Lexer lexer("");
    auto result = lexer.next_token();
    ASSERT_TRUE(result.ok());
    EXPECT_EQ(result.value().type, TokenType::Eof);
}

TEST(LexerTest, Delimiters) {
    Lexer lexer("{}()[],:\"");
    
    auto tok = lexer.next_token();
    ASSERT_TRUE(tok.ok());
    EXPECT_EQ(tok.value().type, TokenType::LBrace);
    
    tok = lexer.next_token();
    ASSERT_TRUE(tok.ok());
    EXPECT_EQ(tok.value().type, TokenType::RBrace);
    
    tok = lexer.next_token();
    ASSERT_TRUE(tok.ok());
    EXPECT_EQ(tok.value().type, TokenType::LParen);
    
    tok = lexer.next_token();
    ASSERT_TRUE(tok.ok());
    EXPECT_EQ(tok.value().type, TokenType::RParen);
    
    tok = lexer.next_token();
    ASSERT_TRUE(tok.ok());
    EXPECT_EQ(tok.value().type, TokenType::LBracket);
    
    tok = lexer.next_token();
    ASSERT_TRUE(tok.ok());
    EXPECT_EQ(tok.value().type, TokenType::RBracket);
    
    tok = lexer.next_token();
    ASSERT_TRUE(tok.ok());
    EXPECT_EQ(tok.value().type, TokenType::Comma);
    
    tok = lexer.next_token();
    ASSERT_TRUE(tok.ok());
    EXPECT_EQ(tok.value().type, TokenType::Colon);
}

TEST(LexerTest, Identifier) {
    Lexer lexer("{name}");
    
    auto tok = lexer.next_token();
    ASSERT_TRUE(tok.ok());
    EXPECT_EQ(tok.value().type, TokenType::LBrace);
    
    tok = lexer.next_token();
    ASSERT_TRUE(tok.ok());
    EXPECT_EQ(tok.value().type, TokenType::Ident);
    EXPECT_EQ(*tok.value().str(), "name");
    
    tok = lexer.next_token();
    ASSERT_TRUE(tok.ok());
    EXPECT_EQ(tok.value().type, TokenType::RBrace);
}

TEST(LexerTest, QuotedString) {
    Lexer lexer(R"("hello world")");
    
    auto tok = lexer.next_token();
    ASSERT_TRUE(tok.ok());
    EXPECT_EQ(tok.value().type, TokenType::Str);
    EXPECT_EQ(*tok.value().str(), "hello world");
}

TEST(LexerTest, EscapedQuotes) {
    Lexer lexer(R"("say \"hi\"")");
    
    auto tok = lexer.next_token();
    ASSERT_TRUE(tok.ok());
    EXPECT_EQ(tok.value().type, TokenType::Str);
    EXPECT_EQ(*tok.value().str(), "say \"hi\"");
}

TEST(LexerTest, Comments) {
    Lexer lexer("/* comment */ {name}");
    
    auto tok = lexer.next_token();
    ASSERT_TRUE(tok.ok());
    EXPECT_EQ(tok.value().type, TokenType::LBrace);
    
    tok = lexer.next_token();
    ASSERT_TRUE(tok.ok());
    EXPECT_EQ(tok.value().type, TokenType::Ident);
}

TEST(LexerTest, Whitespace) {
    Lexer lexer("  {  name  }  ");
    
    auto tok = lexer.next_token();
    ASSERT_TRUE(tok.ok());
    EXPECT_EQ(tok.value().type, TokenType::LBrace);
    
    tok = lexer.next_token();
    ASSERT_TRUE(tok.ok());
    EXPECT_EQ(tok.value().type, TokenType::Ident);
    
    tok = lexer.next_token();
    ASSERT_TRUE(tok.ok());
    EXPECT_EQ(tok.value().type, TokenType::RBrace);
}

TEST(LexerTest, UnclosedComment) {
    Lexer lexer("/* unclosed");
    
    auto tok = lexer.next_token();
    EXPECT_TRUE(tok.is_error());
}

TEST(LexerTest, UnclosedString) {
    Lexer lexer(R"("unclosed)");
    
    auto tok = lexer.next_token();
    EXPECT_TRUE(tok.is_error());
}

TEST(LexerTest, Tokenize) {
    Lexer lexer("{a,b}");
    auto result = lexer.tokenize();
    ASSERT_TRUE(result.ok());
    
    auto& tokens = result.value();
    EXPECT_EQ(tokens.size(), 6); // { a , b } EOF
    EXPECT_EQ(tokens[0].type, TokenType::LBrace);
    EXPECT_EQ(tokens[1].type, TokenType::Ident);
    EXPECT_EQ(tokens[2].type, TokenType::Comma);
    EXPECT_EQ(tokens[3].type, TokenType::Ident);
    EXPECT_EQ(tokens[4].type, TokenType::RBrace);
    EXPECT_EQ(tokens[5].type, TokenType::Eof);
}

