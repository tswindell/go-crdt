/*
 * Copyright (c) 2015 Tom Swindell (t.swindell@rubyx.co.uk)
 *
 * Permission is hereby granted, free of charge, to any person obtaining a copy
 * of this software and associated documentation files (the "Software"), to deal
 * in the Software without restriction, including without limitation the rights
 * to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
 * copies of the Software, and to permit persons to whom the Software is
 * furnished to do so, subject to the following conditions:
 *
 * The above copyright notice and this permission notice shall be included in
 * all copies or substantial portions of the Software.
 *
 * THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
 * IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
 * FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
 * AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
 * LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
 * OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
 * SOFTWARE.
 *
 */

package set

// Common Go representation of a math set.
type GSet map[interface{}]struct{}

func (s *GSet) Insert(item interface{}) bool {
    found := s.Contains(item)
    (*s)[item] = struct{}{}
    return !found
}

func (s *GSet) Contains(item interface{}) bool {
    _, found := (*s)[item]
    return found
}

func (s *GSet) Length() int {
    return len(*s)
}

func (s *GSet) Equals(other GSet) bool {
    if len(*s) != len(other) {
        return false
    }
    for i := range *s {
        if !other.Contains(i) {
            return false
        }
    }
    return true
}

func (s *GSet) Clear() {
    *s = make(GSet)
}

func (s *GSet) Clone() GSet {
    result := make(GSet)
    for i := range *s {
        result.Insert(i)
    }
    return result
}

func (s *GSet) Merge(other GSet) {
    for i := range other {
        s.Insert(i)
    }
}

func (s *GSet) Iterate() <-chan interface{} {
    ch := make(chan interface{})
    go func() {
        for i := range *s { ch <- i }
        close(ch)
    }()
    return ch
}

func (s *GSet) ToSlice() []interface{} {
    result := make([]interface{}, 0, len(*s))
    for i := range *s {
        result = append(result, i)
    }
    return result
}

