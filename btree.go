package main

import (
	"fmt"
	"sort"
)

// B+Treeのノードサイズ（キーの最大数）
// 学習用で小さな値を置いておく
const BTREE_ORDER = 4

// BTreeNodeはB+Treeのノードを表す
// b tree の leaf node は複数の value を持っている（1：1対応ではない）
// leaf node は keys, values を4(BTREE_ORDER)つ持ち、内部ノードは　子ノードを 5 つもつ ( = BTREE_ORDER + 1)
// 内部ノードでは、key = 境界の値, leaf node だと key = values（実際のデータ）
// だからこノード自体は key + 1 になるのか
type BTreeNode struct {
	IsLeaf   bool        // リーフノードかどうか
	Keys     []int       // キーの配列（ソート済み）
	Values   []int       // 値の配列（リーフノードの場合：レコード位置、内部ノードの場合：子ノードのインデックス）
	Children []*BTreeNode // 子ノードへのポインタ（内部ノードのみ）
	Next     *BTreeNode  // 次のリーフノードへのポインタ（リーフノードのみ）
}

// BTreeはB+Treeの根ノードを管理する
type BTree struct {
	Root      *BTreeNode // 根ノード
	TableName string     // 対象テーブル名
	ColumnName string    // 対象カラム名
}

// NewBTree - 新しいB+Treeを作成
func NewBTree(tableName, columnName string) *BTree {
	// 初期状態では空のリーフノードを根とする
	root := &BTreeNode{
		IsLeaf:   true,
		Keys:     make([]int, 0, BTREE_ORDER),
		Values:   make([]int, 0, BTREE_ORDER),
		Children: nil,
		Next:     nil, // leaf node 同士は範囲検索（where）をするために、連結リストで結ばれる
	}
	
	return &BTree{
		Root:       root,
		TableName:  tableName,
		ColumnName: columnName,
	}
}

// Insert - B+Treeにキー・値のペアを挿入
func (bt *BTree) Insert(key int, value int) {
	root := bt.Root
	
	// 根ノードが満杯の場合は分割
	if len(root.Keys) >= BTREE_ORDER {
		newRoot := &BTreeNode{
			IsLeaf:   false,
			Keys:     make([]int, 0, BTREE_ORDER),
			Values:   make([]int, 0, BTREE_ORDER),
			Children: make([]*BTreeNode, 0, BTREE_ORDER+1),
		}
		
		newRoot.Children = append(newRoot.Children, root)
		bt.splitChild(newRoot, 0)
		bt.Root = newRoot
	}
	
	bt.insertNonFull(bt.Root, key, value)
}

// insertNonFull - 満杯でないノードに挿入
func (bt *BTree) insertNonFull(node *BTreeNode, key int, value int) {
	if node.IsLeaf {
		// リーフノードの場合、適切な位置に挿入
		pos := sort.Search(len(node.Keys), func(i int) bool {
			return node.Keys[i] >= key
		})
		
		// 既存キーの場合は値を更新
		if pos < len(node.Keys) && node.Keys[pos] == key {
			node.Values[pos] = value
			return
		}
		
		// 新しいキーを挿入
		node.Keys = append(node.Keys, 0)
		node.Values = append(node.Values, 0)
		
		// 挿入位置を空けるためにシフト
		copy(node.Keys[pos+1:], node.Keys[pos:])
		copy(node.Values[pos+1:], node.Values[pos:])
		
		node.Keys[pos] = key
		node.Values[pos] = value
	} else {
		// 内部ノード( != leaf node ) の場合、適切な子ノードを見つけて再帰的に挿入
		pos := sort.Search(len(node.Keys), func(i int) bool {
			return node.Keys[i] > key
		})
		
		child := node.Children[pos]
		
		// 子ノードが満杯の場合は分割
		if len(child.Keys) >= BTREE_ORDER {
			bt.splitChild(node, pos)
			
			// 分割後、適切な子ノードを選択
			if key > node.Keys[pos] {
				pos++
			}
		}
		
		bt.insertNonFull(node.Children[pos], key, value)
	}
}

// splitChild - 満杯の子ノードを分割
func (bt *BTree) splitChild(parent *BTreeNode, childIndex int) {
	fullChild := parent.Children[childIndex]
	// btree では、真ん中のキーを親に昇格させる
	// この時昇格させるのは、あくまでキーであり、value を持った leaf node ではない
	// 親ノードが保持しているのは、n 以上の値なら右側のリーフに入ってるから(右を分割していく設計なら)そちらを探索しなよ、というキー
	// 子ノードと一緒のキーを持つことになる
	mid := BTREE_ORDER / 2
	
	// 新しいノードを作成（右半分）
	newChild := &BTreeNode{
		IsLeaf:   fullChild.IsLeaf,
		Keys:     make([]int, 0, BTREE_ORDER),
		Values:   make([]int, 0, BTREE_ORDER),
		Children: nil,
		Next:     nil,
	}
	
	if fullChild.IsLeaf {
		// リーフノードの場合
		// leaf node の場合、親ノードを作った上で、リーフノードも作ってるんか
		// つまり、leaf として新しい childNode を生成して、それに元の keys, values の半分をコピー
		// その上で、境界値の key を親ノードに持たせる(昇格させる）のか
		newChild.Keys = append(newChild.Keys, fullChild.Keys[mid:]...)
		newChild.Values = append(newChild.Values, fullChild.Values[mid:]...)
		
		// リンクリストの更新
		newChild.Next = fullChild.Next
		fullChild.Next = newChild
		
		// 元のノードを左半分に縮小
		fullChild.Keys = fullChild.Keys[:mid]
		fullChild.Values = fullChild.Values[:mid]
		
		// 親ノードに中央キーを昇格（リーフの場合は最初のキーをコピー）
		promotedKey := newChild.Keys[0]
		
		// 親ノードの slice をダミー値で拡張
		parent.Keys = append(parent.Keys, 0)
		parent.Children = append(parent.Children, nil)
		
		// 親ノードの中で　promotedKey と value を入れる位置を格納する
		copy(parent.Keys[childIndex+1:], parent.Keys[childIndex:])
		copy(parent.Children[childIndex+2:], parent.Children[childIndex+1:])
		
		parent.Keys[childIndex] = promotedKey
		parent.Children[childIndex+1] = newChild
	} else {
		// 内部ノードの場合
		// 子ノードの分割が必要になる代わりに linked list 周りの処理が不要って感じかなぁ
		newChild.Keys = append(newChild.Keys, fullChild.Keys[mid+1:]...)
		newChild.Values = append(newChild.Values, fullChild.Values[mid+1:]...)
		newChild.Children = append(newChild.Children, fullChild.Children[mid+1:]...)
		
		// 昇格させるキー
		promotedKey := fullChild.Keys[mid]
		
		// 元のノードを左半分に縮小
		fullChild.Keys = fullChild.Keys[:mid]
		fullChild.Values = fullChild.Values[:mid]
		fullChild.Children = fullChild.Children[:mid+1]
		
		// 親ノードに昇格キーと新しい子ノードを挿入
		parent.Keys = append(parent.Keys, 0)
		parent.Children = append(parent.Children, nil)
		
		copy(parent.Keys[childIndex+1:], parent.Keys[childIndex:])
		copy(parent.Children[childIndex+2:], parent.Children[childIndex+1:])
		
		parent.Keys[childIndex] = promotedKey
		parent.Children[childIndex+1] = newChild
	}
}

// Search - B+Treeからキーを検索して値を取得
func (bt *BTree) Search(key int) (int, bool) {
	return bt.searchNode(bt.Root, key)
}

// searchNode - ノード内でキーを検索
func (bt *BTree) searchNode(node *BTreeNode, key int) (int, bool) {
	if node.IsLeaf {
		// リーフノードの場合、線形検索
		pos := sort.Search(len(node.Keys), func(i int) bool {
			return node.Keys[i] >= key
		})
		
		if pos < len(node.Keys) && node.Keys[pos] == key {
			return node.Values[pos], true
		}
		return 0, false
	} else {
		// 内部ノードの場合、適切な子ノードを選択して再帰検索
		pos := sort.Search(len(node.Keys), func(i int) bool {
			return node.Keys[i] > key
		})
		
		// TODO: 
		// 実際のRDBでは、スタックオーバーフローなどのパフォーマンス対策として、loop処理をしてるはず
		return bt.searchNode(node.Children[pos], key)
	}
}

// PrintTree - デバッグ用：B+Treeの構造を表示
func (bt *BTree) PrintTree() {
	fmt.Printf("B+Tree for %s.%s:\n", bt.TableName, bt.ColumnName)
	bt.printNode(bt.Root, 0)
}

// printNode - ノードを再帰的に表示
func (bt *BTree) printNode(node *BTreeNode, depth int) {
	if node == nil {
		return
	}
	
	indent := ""
	for i := 0; i < depth; i++ {
		indent += "  "
	}
	
	if node.IsLeaf {
		fmt.Printf("%sLeaf: Keys=%v, Values=%v\n", indent, node.Keys, node.Values)
	} else {
		fmt.Printf("%sInternal: Keys=%v\n", indent, node.Keys)
		for _, child := range node.Children {
			bt.printNode(child, depth+1)
		}
	}
} 
